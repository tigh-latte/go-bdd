package bdd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/ohler55/ojg/jp"
	"github.com/spf13/viper"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/tigh-latte/go-bdd/bddcontext"
	"github.com/tigh-latte/go-bdd/config"
	"github.com/xlzd/gotp"
	"github.com/zeroflucs-given/generics/collections/stack"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yazgazan/jaydiff/diff"
)

type stepAdder struct {
	ctx StepAdder
}

func (sa *stepAdder) Step(expr, stepFunc any) {
	sa.ctx.Step(expr, func() any {
		if config.IsDryRun() {
			return func() error {
				return godog.ErrSkip
			}
		}
		return stepFunc
	}())
}

// Regex strings.
const (
	REISendARequestTo                 = `^I send a (HEAD|GET|DELETE|POST|PATCH|PUT) request to "([^:/]*)?(:\d+)?([^"]*)"$`
	REISendARequestToWithJSON         = `^I send a (HEAD|GET|DELETE|POST|PATCH|PUT) request to "([^:/]*)?(:\d+)?([^"]*)" with JSON "([^"]*)"$`
	REISendARequestToWithJSONAsString = `^I send a (HEAD|GET|DELETE|POST|PATCH|PUT) request to "([^:/]*)?(:\d+)?([^"]*)" with JSON:$`
)

func initSteps(ctx StepAdder) {
	// S3
	ctx.Step(`^I put the following files into the corresponding s3 buckets:$`, IPutFilesIntoS3)
	ctx.Step(`^there should be (\d+) files in the directory "([^"]*)" in bucket "([^"]*)$"`, ThereShouldBeFilesInDirectoryInBucket)
	ctx.Step(`^the following files should exist in the corresponding s3 buckets:$`, TheFollowingFilesShouldExistInS3Buckets)
	ctx.Step(`I delete the following files from the corresponding s3 buckets:$`, IDeleteFilesFromS3)

	// Mongo
	ctx.Step("^I put the following documents in the corresponding collections and databases:$", IPutDocumentsInMongo)
	ctx.Step("^I put the following documents in the corresponding collections:$", IPutDocumentsInMongoColl)
	ctx.Step(`^the following document IDs should exist in the corresponding mongo collections:$`, TheFollowingDocumentsShouldExistInMongoCollections)
	ctx.Step(`^the following documents should match the following files:$`, TheFollowingDocumentsShouldMatchTheFollowingFiles)
	ctx.Step(`^the database document should match the following values:$`, TheDocumentShouldMatchTheFollowingValues)
	ctx.Step(`^I drop the following mongo databases:$`, IDropMongoDatabase)

	// General
	ctx.Step(`^the headers:$`, TheHeaders)
	ctx.Step(`^I store for templating:$`, IStoreForTemplating)
	ctx.Step(`^I store from the response for templating:$`, IStoreFromTheResponseForTemplating)
	ctx.Step(`^I am unauthenticated$`, IAmUnauthenticated)
	ctx.Step(`^I wait for (\d+) second(s)?$`, IWaitForSeconds)

	// MFA
	ctx.Step(`I create a one time passcode from "([^"]+)"$`, ICreateAOneTimePasscodeFrom)

	// QR code
	ctx.Step(`^I decode the QR code "(.+)"$`, IDecodeTheQRCode)

	// HTTP
	ctx.Step(`^The cookies:$`, TheCookies)
	ctx.Step(`^I set the query params:$`, ISetTheQueryParams)
	ctx.Step(`^I add the query params:$`, IAddTheQueryParams)
	ctx.Step(`^I ignore from all responses:$`, IIgnoreFromAllResponses)
	ctx.Step(REISendARequestTo, ISendARequestTo)
	ctx.Step(REISendARequestToWithJSON, ISendARequestToWithJSON)
	ctx.Step(REISendARequestToWithJSONAsString, ISendARequestToWithJSONAsString)
	ctx.Step(`^the HTTP response code should be (\d*)$`, TheHTTPResponseCodeShouldBe)
	ctx.Step(`^the response should be (\w*)$`, TheResponseShouldBe)
	ctx.Step(`^the response is (\w*)$`, TheResponseShouldBe)
	ctx.Step(`^the response body should match JSON "([^"]*)"$`, TheResponseBodyShouldMatchJSON)
	ctx.Step(`^the response body should match JSON "([^"]*)" ignoring:$`, TheResponseBodyShouldMatchJSONIgnoring)

	// GRPC
	// ctx.Step(`^I send a GRPC request to "([^"]*)"$`, t.ISendAGRPCRequestTo)
	// ctx.Step(`^I send a GRPC request to "([^"]*)" with JSON "([^"]*)"$`, t.ISendAGRPCRequestToWithJSON)

	// Websocket
	// steps support multiple connections per scenario by specifying connection name, else it will use default connection name
	ctx.Step(`^I establish websocket connection$`, IEstablishWebsocketConnection)
	ctx.Step(`^I establish websocket connection called "([^"]*)"$`, IEstablishWebsocketConnectionCalled)
	ctx.Step(`^I send a websocket event "([^"]*)"$`, ISendWebsocketEvent)
	ctx.Step(`^I send a websocket event "([^"]*)" to "([^"]*)" connection$`, ISendWebsocketEventTo)
	ctx.Step(`^I send a websocket event "([^"]*)" with message "([^"]*)"$`, ISendWebsocketEventWithMessage)
	ctx.Step(`^I send a websocket event "([^"]*)" to "([^"]*)" connection with message "([^"]*)"$`, ISendWebsocketEventWithMessageTo)
	ctx.Step(`^I receive websocket event "([^"]*)"$`, IReceiveWebsocketEvent)
	ctx.Step(`^I receive websocket event "([^"]*)" on "([^"]*)" connection$`, IReceiveWebsocketEventTo)
	ctx.Step(`^I wait for websocket event "([^"]*)"$`, IReceiveWebsocketEvent)
	ctx.Step(`^I wait for websocket event "([^"]*)" on "([^"]*)" connection$`, IReceiveWebsocketEventTo)
	ctx.Step(`^the websocket message should match "([^"]*)"$`, TheWebsocketMessageShouldMatchJSON)
	ctx.Step(`^the websocket message on "([^"]*)" connection should match "([^"]*)"$`, TheWebsocketMessageToConnectionShouldMatchJSON)
	ctx.Step(`^the websocket message should match "([^"]*)" ignoring:$`, TheWebsocketMessageShouldMatchJSONIgnoring)
	ctx.Step(`^the websocket message on "([^"]*)" connection should match "([^"]*)" ignoring:$`, TheWebsocketMessageToConnectionShouldMatchJSONIgnoring)
}

func IAmUnauthenticated(ctx context.Context) context.Context {
	return UseAuthentication(ctx, &NoAuthAuthentication{})
}

func IWaitForSeconds(t int, _ string) error {
	tick := time.NewTicker(time.Duration(t) * time.Second)
	defer tick.Stop()
	<-tick.C
	return nil
}

func ICreateAOneTimePasscodeFrom(ctx context.Context, otpUrl string) error {
	t := bddcontext.LoadContext(ctx)
	otpUrl, err := TemplateValue(otpUrl).Render(ctx)
	if err != nil {
		return fmt.Errorf("failed to render otpUrl: %w", err)
	}
	u, err := url.Parse(otpUrl)
	if err != nil {
		return fmt.Errorf("failed to parse otp url %q: %w", otpUrl, err)
	}
	totp := gotp.NewDefaultTOTP(u.Query().Get("secret"))

	code := totp.AtTime(time.Now())
	t.TemplateValues["otp"] = code
	return nil
}

func IDecodeTheQRCode(ctx context.Context, code string) error {
	t := bddcontext.LoadContext(ctx)
	code, err := TemplateValue(code).Render(ctx)
	if err != nil {
		return err
	}

	icon, err := oksvg.ReadIconStream(strings.NewReader(code))
	if err != nil {
		return fmt.Errorf("failed to read icon stream for qrcode: %w", err)
	}
	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)

	bmp, err := gozxing.NewBinaryBitmapFromImage(rgba)
	if err != nil {
		return fmt.Errorf("failed to create binary bitmap: %w", err)
	}
	rdr := qrcode.NewQRCodeReader()
	result, err := rdr.Decode(bmp, nil)
	if err != nil {
		return fmt.Errorf("failed to decode qr code: %w", err)
	}

	if err = t.QRCodes.Push(result); err != nil {
		return fmt.Errorf("failed to add decoded qr code to stack: %w", err)
	}

	t.TemplateValues["qrcode_text"] = result.String()
	return nil
}

// MONGO FUNCTIONS
func IPutDocumentsInMongo(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			db         = row.Cells[0].Value
			collection = row.Cells[1].Value
			file       = row.Cells[2].Value
		)

		r, err := t.MongoContext.TestData.Open(file)
		if err != nil {
			return fmt.Errorf("failed to load file '%s': %w", file, err)
		}

		// Load the collection
		coll := t.MongoContext.Client.Database(db).Collection(collection)

		// Load the documents from the file
		var docs []interface{}
		if err := json.NewDecoder(r).Decode(&docs); err != nil {
			return fmt.Errorf("failed to decode json from file '%s': %w", file, err)
		}

		// Insert the documents
		if _, err := coll.InsertMany(ctx, docs); err != nil {
			return fmt.Errorf("failed to insert documents into collection '%s': %w", collection, err)
		}
	}
	return nil
}

func IPutDocumentsInMongoColl(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			collection = row.Cells[0].Value
			file       = row.Cells[1].Value
		)

		r, err := t.MongoContext.TestData.Open(file)
		if err != nil {
			return fmt.Errorf("failed to load file '%s': %w", file, err)
		}

		// Load the collection
		coll := t.MongoContext.Client.Database(t.ID).Collection(collection)

		// Load the documents from the file
		var docs []interface{}
		if err := json.NewDecoder(r).Decode(&docs); err != nil {
			return fmt.Errorf("failed to decode json from file '%s': %w", file, err)
		}

		// Insert the documents
		if _, err := coll.InsertMany(ctx, docs); err != nil {
			return fmt.Errorf("failed to insert documents into collection '%s': %w", collection, err)
		}
	}
	return nil
}

func TheFollowingDocumentsShouldExistInMongoCollections(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			db         = row.Cells[0].Value
			collection = row.Cells[1].Value
			id         = row.Cells[2].Value
		)

		// Load the collection
		coll := t.MongoContext.Client.Database(db).Collection(collection)

		// Load the document
		var doc interface{}
		if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
			return fmt.Errorf("failed to find document in collection '%s': %w", collection, err)
		}

	}
	return nil
}

func TheDocumentShouldMatchTheFollowingValues(ctx context.Context, docString *godog.DocString) error {
	t := bddcontext.LoadContext(ctx)

	// Pop the ID and collection
	ok, id := t.MongoContext.IDs.Pop()
	if !ok {
		return errors.New("no document ID has been set")
	}

	collection, ok := t.MongoContext.DocumentIDMap[id]
	if !ok {
		return fmt.Errorf("no collection found for document ID '%s'", id)
	}

	// Load the collection
	coll := t.MongoContext.Client.Database(t.ID).Collection(collection)

	// Load the document
	var doc bson.M
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		return fmt.Errorf("failed to find document in collection '%s': %w", collection, err)
	}

	// Marshal the docstring to delete the keys which don't exist
	var expDoc map[string]interface{}
	if err := json.NewDecoder(strings.NewReader(docString.Content)).Decode(&expDoc); err != nil {
		return fmt.Errorf("failed to decode json from docstring: %w", err)
	}

	for k := range doc {
		if _, ok := expDoc[k]; !ok {
			delete(doc, k)
		}
	}

	// Marshal the document to bytes
	b, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// compare the documents
	return compare(ctx, []byte(docString.Content), b, t.MongoContext.ToIgnore, nil)
}

func TheFollowingDocumentsShouldMatchTheFollowingFiles(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			db         = row.Cells[0].Value
			collection = row.Cells[1].Value
			id         = row.Cells[2].Value
			file       = row.Cells[3].Value
		)

		// Load the collection
		coll := t.MongoContext.Client.Database(db).Collection(collection)

		// Load the document
		var doc map[string]interface{}
		if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
			return fmt.Errorf("failed to find document in collection '%s': %w", collection, err)
		}

		// Load the file
		r, err := t.MongoContext.TestData.Open(file)
		if err != nil {
			return fmt.Errorf("failed to load file '%s': %w", file, err)
		}

		// Load the expected document
		var expDoc map[string]interface{}
		if err := json.NewDecoder(r).Decode(&expDoc); err != nil {
			return fmt.Errorf("failed to decode json from file '%s': %w", file, err)
		}

		// Compare the documents
		if !reflect.DeepEqual(doc, expDoc) {
			return fmt.Errorf("documents do not match. expected: %v got: %v", expDoc, doc)
		}

	}
	return nil
}

func IDropMongoDatabase(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		db := row.Cells[0].Value

		if err := t.MongoContext.Client.Database(db).Drop(ctx); err != nil {
			return fmt.Errorf("failed to drop database '%s': %w", db, err)
		}
	}
	return nil
}

func IPutFilesIntoS3(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			bucket = row.Cells[0].Value
			file   = row.Cells[1].Value
			key    = row.Cells[2].Value
		)

		r, err := t.TestData.Open(file)
		if err != nil {
			return fmt.Errorf("failed to load file '%s': %w", file, err)
		}
		if _, err = t.S3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Body:   r,
			Key:    aws.String(key),
		}); err != nil {
			return fmt.Errorf("failed to put file '%s': %w", file, err)
		}
	}
	return nil
}

func IDeleteFilesFromS3(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			bucket = row.Cells[0].Value
			prefix = row.Cells[1].Value
		)

		list, err := t.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(prefix),
		})
		if err != nil {
			return fmt.Errorf("failed to list files in bucket '%s' with prefix '%s' to be deleted: %w", bucket, prefix, err)
		}

		objects := make([]types.ObjectIdentifier, len(list.Contents))
		for i, v := range list.Contents {
			objects[i] = types.ObjectIdentifier{
				Key: v.Key,
			}
		}

		if len(objects) == 0 {
			return nil
		}

		if _, err = t.S3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{
				Objects: objects,
			},
		}); err != nil {
			return fmt.Errorf("failed to delete from bucket '%s' with prefix '%s': %w", bucket, prefix, err)
		}
	}
	return nil
}

func ThereShouldBeFilesInDirectoryInBucket(ctx context.Context, exp int32, key, bucket string) error {
	t := bddcontext.LoadContext(ctx)
	p := s3.NewListObjectsV2Paginator(t.S3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})

	var (
		total     int32
		pageCount = 1
	)
	for ; p.HasMorePages(); pageCount++ {
		page, err := p.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("error fetching page %d in bucket '%s' with prefix '%s': %w", pageCount, bucket, key, err)
		}
		total += *page.KeyCount
	}

	if exp != total {
		return fmt.Errorf("incorrect count for bucket '%s' dir '%s'. expected: %d got: %d", bucket, key, exp, total)
	}

	return nil
}

func TheFollowingFilesShouldExistInS3Buckets(ctx context.Context, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			bucket = row.Cells[0].Value
			key    = row.Cells[1].Value
		)

		if _, err := t.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}); err != nil {
			var aerr *awshttp.ResponseError
			if errors.As(err, &aerr) && aerr.ResponseError.HTTPStatusCode() == http.StatusNotFound {
				return fmt.Errorf("file '%s/%s' does not exist: %w", bucket, key, err)
			}
			return fmt.Errorf("failed to check if file '%s/%s' exists: %w", bucket, key, err)
		}
	}

	return nil
}

func TheHeaders(ctx context.Context, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			key, kErr = TemplateValue(row.Cells[0].Value).Render(ctx)
			val, vErr = TemplateValue(row.Cells[1].Value).Render(ctx)
		)
		if kErr != nil {
			return ctx, fmt.Errorf("failed to render header key %q: %w", key, kErr)
		}
		if vErr != nil {
			return ctx, fmt.Errorf("failed to render header value %q: %w", val, vErr)
		}
		t.HTTP.Headers.Set(key, val)
	}

	return bddcontext.WithContext(ctx, t), nil
}

func IStoreForTemplating(ctx context.Context, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			jpath = row.Cells[0].Value
			file  = row.Cells[1].Value
		)

		tmpl, err := t.Template.Parse(file)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse template for file '%s': %w", file, err)
		}
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, t.TemplateValues); err != nil {
			return ctx, fmt.Errorf("failed to execute template for file '%s': %w", file, err)
		}
		x := jp.MustParseString("$" + jpath)
		var value any = buf.String()

		if err = x.Set(t.TemplateValues, value); err != nil {
			return ctx, fmt.Errorf("failed to set template path '%s': %w", jpath, err)
		}
	}

	return bddcontext.WithContext(ctx, t), nil
}

func IStoreFromTheResponseForTemplating(ctx context.Context, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)
	ok, resp := t.HTTP.Responses.Peek()
	if !ok {
		return ctx, errors.New("no http request has been made yet")
	}

	return doTemplate(ctx, resp, table)
}

func doTemplate(ctx context.Context, data []byte, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		var (
			jpath = row.Cells[0].Value
			as    = row.Cells[1].Value
		)

		var jsonData any
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return ctx, err
		}

		x := jp.MustParseString("$" + jpath)
		vals := x.Get(jsonData)
		if len(vals) == 1 {
			t.TemplateValues[as] = vals[0]
			continue
		}

		for i, val := range vals {
			t.TemplateValues[fmt.Sprintf("%s_%d", as, i)] = val
		}
	}

	return bddcontext.WithContext(ctx, t), nil
}

func TheCookies(ctx context.Context, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	for _, row := range table.Rows[1:] {
		var (
			key   = row.Cells[0].Value
			value = row.Cells[1].Value
		)

		tmpl, err := t.Template.Parse(value)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse template for value '%s': %w", value, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, t.TemplateValues); err != nil {
			return ctx, fmt.Errorf("failed ot execute template for value '%s': %w", value, err)
		}

		t.HTTP.Cookies = append(t.HTTP.Cookies, &http.Cookie{
			Name:    key,
			Value:   value,
			Expires: time.Now().Add(time.Minute * 60),
		})
	}

	return bddcontext.WithContext(ctx, t), nil
}

func ISetTheQueryParams(ctx context.Context, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	for _, row := range table.Rows[1:] {
		var (
			key   = row.Cells[0].Value
			value = row.Cells[1].Value
		)

		tmpl, err := t.Template.Parse(value)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse template for value '%s': %w", value, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, t.TemplateValues); err != nil {
			return ctx, fmt.Errorf("failed to execute template for value '%s': '%w'", value, err)
		}

		t.HTTP.QueryParams.Set(key, buf.String())
	}

	return bddcontext.WithContext(ctx, t), nil
}

func IAddTheQueryParams(ctx context.Context, table *godog.Table) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	for _, row := range table.Rows[1:] {
		var (
			key   = row.Cells[0].Value
			value = row.Cells[1].Value
		)

		tmpl, err := t.Template.Parse(value)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse template for value '%s': %w", value, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, t.TemplateValues); err != nil {
			return ctx, fmt.Errorf("failed to execute template for value '%s': '%w'", value, err)
		}

		t.HTTP.QueryParams.Add(key, buf.String())
	}

	return bddcontext.WithContext(ctx, t), nil
}

func IIgnoreFromAllResponses(ctx context.Context, table *godog.Table) context.Context {
	t := bddcontext.LoadContext(ctx)
	for _, row := range table.Rows[1:] {
		t.HTTP.ToIgnore = append(t.HTTP.ToIgnore, row.Cells[0].Value)
	}

	return bddcontext.WithContext(ctx, t)
}

func ISendARequestTo(ctx context.Context, verb, host, port, endpoint string) (context.Context, error) {
	return ISendARequestToWithJSON(ctx, verb, host, port, endpoint, "")
}

func ISendARequestToWithJSONAsString(ctx context.Context, verb, host, port, endpoint, payload string) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	uriPath, qp, err := func() (string, string, error) {
		tmpl, err := t.Template.Parse(endpoint)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse template for endpoint '%s': %w", endpoint, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, t.TemplateValues); err != nil {
			return "", "", fmt.Errorf("failed to execute template for endpoint '%s': %w", endpoint, err)
		}

		parts := strings.Split(buf.String(), "?")
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		u := url.URL{RawQuery: parts[1]}
		qp := u.Query()
		for k := range t.HTTP.QueryParams {
			qp.Add(k, t.HTTP.QueryParams.Get(k))
		}
		return parts[0], qp.Encode(), nil
	}()
	if err != nil {
		return ctx, err
	}
	if host == "" {
		host = viper.GetString("service.url")
	} else {
		host = "https://" + host
	}
	url, err := url.Parse(host)
	if err != nil {
		return ctx, fmt.Errorf("failed to url parse service href: %w", err)
	}
	url.Path = uriPath
	url.RawQuery = qp

	req, err := http.NewRequest(verb, url.String(), nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to make request to '%s': %w", url.String(), err)
	}

	req.Header.Add("Content-Type", "application/json")

	for k, v := range t.HTTP.Headers {
		for _, header := range v {
			req.Header.Add(k, header)
		}
	}
	for _, cookie := range t.HTTP.Cookies {
		req.AddCookie(cookie)
	}

	if auth, ok := GetAuthentication(ctx); ok {
		auth.ApplyHTTP(ctx, req)
	}

	bb, err := func() ([]byte, error) {
		if payload != "" {
			payload, err = TemplateValue(payload).Render(ctx)
			if err != nil {
				return []byte{}, fmt.Errorf("failed to JSON body: %w", err)
			}
			if !json.Valid([]byte(payload)) {
				return []byte{}, nil
			}
			return []byte(payload), nil
		}

		return []byte{}, nil
	}()
	if err != nil {
		return ctx, err
	}

	if err := t.HTTP.Requests.Push(bb); err != nil {
		return ctx, fmt.Errorf("you've flown too close to the sun: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewBuffer(bb))

	resp, err := t.HTTP.Client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to do request to '%s': %w", url.String(), err)
	}

	defer resp.Body.Close()

	if err := t.HTTP.ResponseCodes.Push(resp.StatusCode); err != nil {
		return ctx, fmt.Errorf("you've flown too close to the sun: %w", err)
	}

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response from '%s': %w", url.String(), err)
	}
	if len(rawResp) == 0 {
		rawResp = []byte{'{', '}'}
	}

	if err := t.HTTP.Responses.Push(rawResp); err != nil {
		return ctx, fmt.Errorf("you've flown too close to the sun: %w", err)
	}

	ctx = bddcontext.WithContext(ctx, t)

	return ctx, nil
}

func ISendARequestToWithJSON(ctx context.Context, verb, host, _, endpoint, file string) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)
	if file != "" {
		file = file + ".json"
	}
	uriPath, qp, err := func() (string, string, error) {
		tmpl, err := t.Template.Parse(endpoint)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse template for endpoint '%s': %w", endpoint, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, t.TemplateValues); err != nil {
			return "", "", fmt.Errorf("failed to execute template for endpoint '%s': %w", endpoint, err)
		}

		parts := strings.Split(buf.String(), "?")
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		u := url.URL{RawQuery: parts[1]}
		qp := u.Query()
		for k := range t.HTTP.QueryParams {
			qp.Add(k, t.HTTP.QueryParams.Get(k))
		}
		return parts[0], qp.Encode(), nil
	}()
	if err != nil {
		return ctx, err
	}
	if host == "" {
		host = viper.GetString("service.url")
	} else {
		host = "https://" + host
	}
	url, err := url.Parse(host)
	if err != nil {
		return ctx, fmt.Errorf("failed to url parse service href: %w", err)
	}
	url.Path = uriPath
	url.RawQuery = qp

	bb, err := func() ([]byte, error) {
		if file == "" {
			return []byte{}, nil
		}
		tmpl := template.Must(t.Template.ParseFS(t.HTTP.TestData, path.Join("requests", file)))

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, path.Base(file), t.TemplateValues); err != nil {
			return nil, fmt.Errorf("failed to parse execute for file '%s': %w", file, err)
		}

		return buf.Bytes(), nil
	}()
	if err != nil {
		return ctx, err
	}

	if err := t.HTTP.Requests.Push(bb); err != nil {
		return ctx, fmt.Errorf("you've flown too close to the sun: %w", err)
	}

	req, err := http.NewRequest(verb, url.String(), bytes.NewBuffer(bb))
	if err != nil {
		return ctx, fmt.Errorf("failed to make request to '%s': %w", url.String(), err)
	}

	req.Header.Add("Content-Type", "application/json")

	for k, v := range t.HTTP.Headers {
		for _, header := range v {
			req.Header.Add(k, header)
		}
	}
	for _, cookie := range t.HTTP.Cookies {
		req.AddCookie(cookie)
	}

	if auth, ok := GetAuthentication(ctx); ok {
		auth.ApplyHTTP(ctx, req)
	}

	resp, err := t.HTTP.Client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to do request to '%s': %w", url.String(), err)
	}

	defer resp.Body.Close()

	if err := t.HTTP.ResponseCodes.Push(resp.StatusCode); err != nil {
		return ctx, fmt.Errorf("you've flown too close to the sun: %w", err)
	}

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response from '%s': %w", url.String(), err)
	}
	if len(rawResp) == 0 {
		rawResp = []byte{'{', '}'}
	}

	if err := t.HTTP.Responses.Push(rawResp); err != nil {
		return ctx, fmt.Errorf("you've flown too close to the sun: %w", err)
	}

	ctx = bddcontext.WithContext(ctx, t)

	return ctx, nil
}

func TheHTTPResponseCodeShouldBe(ctx context.Context, code int) error {
	t := bddcontext.LoadContext(ctx)

	ok, lastRespCode := t.HTTP.ResponseCodes.Peek()
	if !ok {
		return errors.New("no http request has been made")
	}
	if code == lastRespCode {
		return nil
	}

	ok, resp := t.HTTP.Responses.Peek()
	if !ok {
		return errors.New("no http request has been made")
	}

	respData := func() []byte {
		if len(resp) < 2000 {
			return make([]byte, len(resp))
		}
		return make([]byte, 2000)
	}()
	copy(respData, resp)

	return fmt.Errorf(
		"expected http response code. wanted %d, got %d with body '%s' using '%s'",
		code, lastRespCode, string(respData), string(resp))
}

func TheResponseShouldBe(ctx context.Context, response string) error {
	return TheHTTPResponseCodeShouldBe(ctx, map[string]int{
		"OK":           http.StatusOK,
		"CREATED":      http.StatusCreated,
		"NO_CONTENT":   http.StatusNoContent,
		"BAD_REQUEST":  http.StatusBadRequest,
		"FORBIDDEN":    http.StatusForbidden,
		"NOT_FOUND":    http.StatusNotFound,
		"UNAUTHORIZED": http.StatusUnauthorized,
		"CONFLICT":     http.StatusConflict,
	}[strings.ToUpper(response)])
}

func TheResponseBodyShouldMatchJSON(ctx context.Context, f string) error {
	return TheResponseBodyShouldMatchJSONIgnoring(ctx, f, &godog.Table{Rows: []*messages.PickleTableRow{{}}})
}

func TheResponseBodyShouldMatchJSONIgnoring(ctx context.Context, f string, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	f = f + ".json"

	tmpl := template.Must(t.Template.ParseFS(t.HTTP.TestData, path.Join("responses", f)))

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, path.Base(f), t.TemplateValues); err != nil {
		return fmt.Errorf("failed to execute template for file '%s': %w", f, err)
	}

	ok, resp := t.HTTP.Responses.Peek()
	if !ok {
		return errors.New("no http request has been made")
	}

	return compare(ctx, buf.Bytes(), resp, t.HTTP.ToIgnore, table)
}

func compare(ctx context.Context, l, r []byte, scenarioIgnore []string, stepIgnore *godog.Table) error {
	t := bddcontext.LoadContext(ctx)
	var exp, got any

	if err := json.Unmarshal(l, &exp); err != nil {
		return fmt.Errorf("failed to marshal expected json '%s': %w", string(l), err)
	}
	if err := json.Unmarshal(r, &got); err != nil {
		return fmt.Errorf("failed to marshal received json '%s': %w", string(r), err)
	}

	// Test suite wide ignores
	for _, entry := range t.IgnoreAlways {
		x := jp.MustParseString("$" + entry)
		if err := x.Del(got); err != nil {
			return fmt.Errorf("failed to remove '%s': %w", entry, err)
		}
	}

	// Request specific ignores
	for _, entry := range scenarioIgnore {
		x := jp.MustParseString("$" + entry)
		if err := x.Del(got); err != nil {
			return fmt.Errorf("failed to remove '%s': %w", entry, err)
		}
	}
	// Scenario specific ignores
	if stepIgnore != nil {
		for _, row := range stepIgnore.Rows[1:] {
			jpath := row.Cells[0].Value

			x := jp.MustParseString("$" + jpath)
			if err := x.Del(got); err != nil {
				return fmt.Errorf("failed to remove '%s': %w", jpath, err)
			}
		}
	}

	d, err := diff.Diff(exp, got)
	if err != nil {
		return fmt.Errorf("error calculating diff: %w", err)
	}
	if d.Diff() == diff.Identical {
		return nil
	}
	report := d.StringIndent("", "", diff.Output{
		Indent:     "  ",
		ShowTypes:  true,
		JSON:       true,
		JSONValues: true,
	})

	return fmt.Errorf(report)
}

func IEstablishWebsocketConnection(ctx context.Context) (context.Context, error) {
	return IEstablishWebsocketConnectionCalled(ctx, "__default")
}

func IEstablishWebsocketConnectionCalled(ctx context.Context, connectionID string) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	if connectionID == "" {
		return ctx, errors.New("missing connection ID")
	}

	if t.WS.Host == "" {
		return ctx, errors.New("websocket host is missing")
	}

	if _, ok := t.WS.Connections[connectionID]; ok {
		return ctx, errors.New("connection ID already exists")
	}

	sid, err := t.WS.Client.Dial(ctx, t.WS.Host)
	if err != nil {
		return ctx, err
	}

	t.WS.Connections[connectionID] = struct {
		SessionID string
		Messages  *stack.Stack[[]byte]
	}{
		SessionID: sid,
		Messages:  stack.NewStack[[]byte](20),
	}

	return bddcontext.WithContext(ctx, t), nil
}

func ISendWebsocketEvent(ctx context.Context, event string) (context.Context, error) {
	return ISendWebsocketEventWithMessageTo(ctx, event, "__default", "")
}

func ISendWebsocketEventTo(ctx context.Context, event, connectionID string) (context.Context, error) {
	return ISendWebsocketEventWithMessageTo(ctx, event, connectionID, "")
}

func ISendWebsocketEventWithMessage(ctx context.Context, event, file string) (context.Context, error) {
	return ISendWebsocketEventWithMessageTo(ctx, event, "__default", file)
}

func ISendWebsocketEventWithMessageTo(ctx context.Context, event, connectionID, file string) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	if connectionID == "" {
		return ctx, errors.New("missing connection ID, please specify websocket connection")
	}

	connection, ok := t.WS.Connections[connectionID]
	if !ok {
		return ctx, fmt.Errorf("specified %s connection is not establish", connectionID)
	}

	if connection.SessionID == "" {
		return ctx, fmt.Errorf("session ID is missing on %s connection", connectionID)
	}

	bb, err := func() ([]byte, error) {
		if file == "" {
			return []byte{}, nil
		}
		tmpl := template.Must(t.Template.ParseFS(t.WS.TestData, path.Join("requests", file)))

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, path.Base(file), t.TemplateValues); err != nil {
			return nil, fmt.Errorf("failed to parse execute for file '%s': %w", file, err)
		}

		return buf.Bytes(), nil
	}()
	if err != nil {
		return ctx, err
	}

	if err := t.WS.Client.Write(ctx, connection.SessionID, event, bb); err != nil {
		return ctx, err
	}

	return bddcontext.WithContext(ctx, t), nil
}

func IReceiveWebsocketEvent(ctx context.Context, event string) (context.Context, error) {
	return IReceiveWebsocketEventTo(ctx, event, "__default")
}

func IReceiveWebsocketEventTo(ctx context.Context, event, connectionID string) (context.Context, error) {
	t := bddcontext.LoadContext(ctx)

	if connectionID == "" {
		return ctx, errors.New("missing connection ID, please specify websocket connection")
	}

	connection, ok := t.WS.Connections[connectionID]
	if !ok {
		return ctx, fmt.Errorf("specified %s connection is not establish", connectionID)
	}

	if connection.SessionID == "" {
		return ctx, fmt.Errorf("session ID is missing on %s connection", connectionID)
	}

	timer := time.NewTimer(t.WS.Timeout)

	ch, err := t.WS.Client.Read(ctx, connection.SessionID, event)
	if err != nil {
		return ctx, err
	}

	select {
	case e := <-ch:
		bb, err := json.MarshalIndent(e.Message, "", "   ")
		if err != nil {
			return ctx, fmt.Errorf("failed to marshal message '%+v'", e.Message)
		}

		connection.Messages.Push(bb)
		t.WS.Connections[connectionID] = connection
	case <-timer.C:
		return ctx, fmt.Errorf("timed out receiving websocket event '%s'", event)
	}

	timer.Stop()

	return bddcontext.WithContext(ctx, t), nil
}

func TheWebsocketMessageShouldMatchJSON(ctx context.Context, f string) error {
	return TheWebsocketMessageToConnectionShouldMatchJSONIgnoring(ctx, "__default", f, &godog.Table{Rows: []*messages.PickleTableRow{{}}})
}

func TheWebsocketMessageToConnectionShouldMatchJSON(ctx context.Context, connectionID, f string) error {
	return TheWebsocketMessageToConnectionShouldMatchJSONIgnoring(ctx, connectionID, f, &godog.Table{Rows: []*messages.PickleTableRow{{}}})
}

func TheWebsocketMessageShouldMatchJSONIgnoring(ctx context.Context, f string, table *godog.Table) error {
	return TheWebsocketMessageToConnectionShouldMatchJSONIgnoring(ctx, "__default", f, table)
}

func TheWebsocketMessageToConnectionShouldMatchJSONIgnoring(ctx context.Context, connectionID, f string, table *godog.Table) error {
	t := bddcontext.LoadContext(ctx)

	if connectionID == "" {
		return errors.New("missing connection ID, please specify websocket connection")
	}

	connection, ok := t.WS.Connections[connectionID]
	if !ok {
		return fmt.Errorf("specified %s connection is not establish", connectionID)
	}

	tmpl := template.Must(t.Template.ParseFS(t.WS.TestData, path.Join("responses", f)))

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, path.Base(f), t.TemplateValues); err != nil {
		return fmt.Errorf("failed to execute template for file '%s': %w", f, err)
	}

	ok, msg := connection.Messages.Peek()
	if !ok {
		return errors.New("no messages received")
	}

	return compare(ctx, buf.Bytes(), msg, t.HTTP.ToIgnore, table)
}
