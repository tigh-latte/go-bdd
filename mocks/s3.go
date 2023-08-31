// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tigh-latte/go-bdd"
	"sync"
)

// Ensure, that S3Mock does implement bdd.S3.
// If this is not the case, regenerate this file with moq.
var _ bdd.S3 = &S3Mock{}

// S3Mock is a mock implementation of bdd.S3.
//
//	func TestSomethingThatUsesS3(t *testing.T) {
//
//		// make and configure a mocked bdd.S3
//		mockedS3 := &S3Mock{
//			DeleteObjectsFunc: func(contextMoqParam context.Context, deleteObjectsInput *s3.DeleteObjectsInput, fns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
//				panic("mock out the DeleteObjects method")
//			},
//			HeadObjectFunc: func(contextMoqParam context.Context, headObjectInput *s3.HeadObjectInput, fns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
//				panic("mock out the HeadObject method")
//			},
//			ListObjectsV2Func: func(contextMoqParam context.Context, listObjectsV2Input *s3.ListObjectsV2Input, fns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
//				panic("mock out the ListObjectsV2 method")
//			},
//			PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
//				panic("mock out the PutObject method")
//			},
//		}
//
//		// use mockedS3 in code that requires bdd.S3
//		// and then make assertions.
//
//	}
type S3Mock struct {
	// DeleteObjectsFunc mocks the DeleteObjects method.
	DeleteObjectsFunc func(contextMoqParam context.Context, deleteObjectsInput *s3.DeleteObjectsInput, fns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)

	// HeadObjectFunc mocks the HeadObject method.
	HeadObjectFunc func(contextMoqParam context.Context, headObjectInput *s3.HeadObjectInput, fns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)

	// ListObjectsV2Func mocks the ListObjectsV2 method.
	ListObjectsV2Func func(contextMoqParam context.Context, listObjectsV2Input *s3.ListObjectsV2Input, fns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)

	// PutObjectFunc mocks the PutObject method.
	PutObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

	// calls tracks calls to the methods.
	calls struct {
		// DeleteObjects holds details about calls to the DeleteObjects method.
		DeleteObjects []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// DeleteObjectsInput is the deleteObjectsInput argument value.
			DeleteObjectsInput *s3.DeleteObjectsInput
			// Fns is the fns argument value.
			Fns []func(*s3.Options)
		}
		// HeadObject holds details about calls to the HeadObject method.
		HeadObject []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// HeadObjectInput is the headObjectInput argument value.
			HeadObjectInput *s3.HeadObjectInput
			// Fns is the fns argument value.
			Fns []func(*s3.Options)
		}
		// ListObjectsV2 holds details about calls to the ListObjectsV2 method.
		ListObjectsV2 []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// ListObjectsV2Input is the listObjectsV2Input argument value.
			ListObjectsV2Input *s3.ListObjectsV2Input
			// Fns is the fns argument value.
			Fns []func(*s3.Options)
		}
		// PutObject holds details about calls to the PutObject method.
		PutObject []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.PutObjectInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
	}
	lockDeleteObjects sync.RWMutex
	lockHeadObject    sync.RWMutex
	lockListObjectsV2 sync.RWMutex
	lockPutObject     sync.RWMutex
}

// DeleteObjects calls DeleteObjectsFunc.
func (mock *S3Mock) DeleteObjects(contextMoqParam context.Context, deleteObjectsInput *s3.DeleteObjectsInput, fns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	if mock.DeleteObjectsFunc == nil {
		panic("S3Mock.DeleteObjectsFunc: method is nil but S3.DeleteObjects was just called")
	}
	callInfo := struct {
		ContextMoqParam    context.Context
		DeleteObjectsInput *s3.DeleteObjectsInput
		Fns                []func(*s3.Options)
	}{
		ContextMoqParam:    contextMoqParam,
		DeleteObjectsInput: deleteObjectsInput,
		Fns:                fns,
	}
	mock.lockDeleteObjects.Lock()
	mock.calls.DeleteObjects = append(mock.calls.DeleteObjects, callInfo)
	mock.lockDeleteObjects.Unlock()
	return mock.DeleteObjectsFunc(contextMoqParam, deleteObjectsInput, fns...)
}

// DeleteObjectsCalls gets all the calls that were made to DeleteObjects.
// Check the length with:
//
//	len(mockedS3.DeleteObjectsCalls())
func (mock *S3Mock) DeleteObjectsCalls() []struct {
	ContextMoqParam    context.Context
	DeleteObjectsInput *s3.DeleteObjectsInput
	Fns                []func(*s3.Options)
} {
	var calls []struct {
		ContextMoqParam    context.Context
		DeleteObjectsInput *s3.DeleteObjectsInput
		Fns                []func(*s3.Options)
	}
	mock.lockDeleteObjects.RLock()
	calls = mock.calls.DeleteObjects
	mock.lockDeleteObjects.RUnlock()
	return calls
}

// HeadObject calls HeadObjectFunc.
func (mock *S3Mock) HeadObject(contextMoqParam context.Context, headObjectInput *s3.HeadObjectInput, fns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if mock.HeadObjectFunc == nil {
		panic("S3Mock.HeadObjectFunc: method is nil but S3.HeadObject was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		HeadObjectInput *s3.HeadObjectInput
		Fns             []func(*s3.Options)
	}{
		ContextMoqParam: contextMoqParam,
		HeadObjectInput: headObjectInput,
		Fns:             fns,
	}
	mock.lockHeadObject.Lock()
	mock.calls.HeadObject = append(mock.calls.HeadObject, callInfo)
	mock.lockHeadObject.Unlock()
	return mock.HeadObjectFunc(contextMoqParam, headObjectInput, fns...)
}

// HeadObjectCalls gets all the calls that were made to HeadObject.
// Check the length with:
//
//	len(mockedS3.HeadObjectCalls())
func (mock *S3Mock) HeadObjectCalls() []struct {
	ContextMoqParam context.Context
	HeadObjectInput *s3.HeadObjectInput
	Fns             []func(*s3.Options)
} {
	var calls []struct {
		ContextMoqParam context.Context
		HeadObjectInput *s3.HeadObjectInput
		Fns             []func(*s3.Options)
	}
	mock.lockHeadObject.RLock()
	calls = mock.calls.HeadObject
	mock.lockHeadObject.RUnlock()
	return calls
}

// ListObjectsV2 calls ListObjectsV2Func.
func (mock *S3Mock) ListObjectsV2(contextMoqParam context.Context, listObjectsV2Input *s3.ListObjectsV2Input, fns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if mock.ListObjectsV2Func == nil {
		panic("S3Mock.ListObjectsV2Func: method is nil but S3.ListObjectsV2 was just called")
	}
	callInfo := struct {
		ContextMoqParam    context.Context
		ListObjectsV2Input *s3.ListObjectsV2Input
		Fns                []func(*s3.Options)
	}{
		ContextMoqParam:    contextMoqParam,
		ListObjectsV2Input: listObjectsV2Input,
		Fns:                fns,
	}
	mock.lockListObjectsV2.Lock()
	mock.calls.ListObjectsV2 = append(mock.calls.ListObjectsV2, callInfo)
	mock.lockListObjectsV2.Unlock()
	return mock.ListObjectsV2Func(contextMoqParam, listObjectsV2Input, fns...)
}

// ListObjectsV2Calls gets all the calls that were made to ListObjectsV2.
// Check the length with:
//
//	len(mockedS3.ListObjectsV2Calls())
func (mock *S3Mock) ListObjectsV2Calls() []struct {
	ContextMoqParam    context.Context
	ListObjectsV2Input *s3.ListObjectsV2Input
	Fns                []func(*s3.Options)
} {
	var calls []struct {
		ContextMoqParam    context.Context
		ListObjectsV2Input *s3.ListObjectsV2Input
		Fns                []func(*s3.Options)
	}
	mock.lockListObjectsV2.RLock()
	calls = mock.calls.ListObjectsV2
	mock.lockListObjectsV2.RUnlock()
	return calls
}

// PutObject calls PutObjectFunc.
func (mock *S3Mock) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if mock.PutObjectFunc == nil {
		panic("S3Mock.PutObjectFunc: method is nil but S3.PutObject was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.PutObjectInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockPutObject.Lock()
	mock.calls.PutObject = append(mock.calls.PutObject, callInfo)
	mock.lockPutObject.Unlock()
	return mock.PutObjectFunc(ctx, params, optFns...)
}

// PutObjectCalls gets all the calls that were made to PutObject.
// Check the length with:
//
//	len(mockedS3.PutObjectCalls())
func (mock *S3Mock) PutObjectCalls() []struct {
	Ctx    context.Context
	Params *s3.PutObjectInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.PutObjectInput
		OptFns []func(*s3.Options)
	}
	mock.lockPutObject.RLock()
	calls = mock.calls.PutObject
	mock.lockPutObject.RUnlock()
	return calls
}
