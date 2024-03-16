# emcoerror package


- [Overview](#overview)
- [Getting started](#getting-started)
- [Usage](#usage)
    - [Handling API errors](#handling-api-errors)
    - [Handling life-cycle errors](#handling-life-cycle-errors)
    - [Managing anticipated errors](#managing-anticipated-errors)
- [Defining error messages](#defining-error-messages)


## Overview

Package `emcoerror` standardizes the error handling process for the controllers in emco. The emcoerror consists of the following.

- `Error` - Type Error implements the emcoerror.
- `ErrorReason` -  The ErrorReason is an enumeration of potential failure reasons. Each emco error type must be associated with an ErrorReason.


## Getting started

The `emcoerror` package defines a set of possible error causes. This list needs to be updated based on the requirements. At any point, we expect a controller should be able to map any failures to one of these reasons. 

``` sh
    const (
        BadRequest ErrorReason = iota
	    Conflict
	    NotFound
	    Unknown
	    // Define other reasons as required
    )
```

Type `Error` implements the emcoerror.

``` sh
    type Error struct {
        error
        Message string
        Reason  ErrorReason
        Cause   *Error
        // Add any additional data
    }
```

The emco error type consists of the following.

- `Message` - The message conveys an explanation of the encountered error.
- `Reason` - This is the failure reason. Each emco error type must be associated with an `ErrorReason`. 
- `Cause` - The cause represents the underlying reason for the error. The error message does not always provide enough information about the root cause of the failure. It is difficult to debug the root cause in cases where we have several calling stacks. The cause can be handy in these scenarios. We can send meaningful responses back to the user, and at the same time, we can use the details in the cause to identify the root cause of the failure. The cause can help us trace the error with comprehensive data throughout the call stack. The cause also avoids the need to wrap failures in each of the call stacks.    

The emco error type can be further extended to include any other details as required.


## Usage

The usage of emco error type is possible anywhere in the stack. Using the emco error reason/ cause, we can take any decisions on the execution flow at any stage. The emco error currently supports the following use cases. We can extend the emcoerror package to handle any future requirements.


### Handling API errors

One of the challenges in error handling is to give the proper HTTP status code to the client/ user along with a meaningful message. Each emco error `reason` must map to an HTTP status code. The emcoerror package has a map of the `ErrorReason` and the HTTP status code. 
 
``` sh
    var StatusCode = map[ErrorReason]int{
	// 4xx
	BadRequest: http.StatusBadRequest,
	Conflict:   http.StatusConflict,
	NotFound:   http.StatusNotFound,
	// 5xx
	Unknown: http.StatusInternalServerError,
    }
```

> **NOTE**: This map must be updated with any new reason or HTTP status code. If the emco error reason does not have a corresponding HTTP status code mapped then the emcoerror package will return the internal server error.

The API handler functions are responsible for sending the appropriate response and HTTP status code to the client/ user. The emcoerror package implements methods to format the response and HTTP status code in case of an API request failure. The error can be an emco error, go built-in error, or other custom-defined error. For example, if the user calls the API to get the caCert and the caCert does not exist in the emco db, the module will return the emco error with the message as `caCert not found`. Once the API handler receives this error, the handler will use the emcoerror package to format the response and the HTTP status code. The emco error type in this example has the reason as `NotFound`. This `ErrorReason` is mapped with the HTTP status code `StatusNotFound`. The emco error type also implements the error interface. If the emco error has a cause defined then the response will be a combination of the message and the cause. The emcoerror package then returns an `APIError` type which contains the response and the HTTP status. The handler will then return the API response to the client with these details.

``` sh
    func (h *cpCertHandler) handleCertificateGet(w http.ResponseWriter, r *http.Request) {
        certs, err = h.manager.GetCert(vars.cert, vars.clusterProvider)
        if err != nil {
            apiErr := emcoerror.HandleAPIError(err)
            http.Error(w, apiErr.Message, apiErr.Status)
            return
        }
    }

    func (c *CaCertClient) GetCert() (CaCert, error) {
        return CaCert{}, &emcoerror.Error{
            Message: CaCertNotFound,
            Reason:  emcoerror.NotFound,
        }
    }

    func HandleAPIError(err error) APIError {
        if status, ok := HTTPStatusCode[e.Reason]; ok {
			return APIError{Message: e.Error(), Status: status}
		}
    }

    // Error implements the error interface
    func (e *Error) Error() string {
        if e.Cause != nil {
            return e.Message + e.Cause.Error()
        }
        return e.Message
    }
```

There are cases when we may want to manage more than one error for a request. We should be able to capture these errors at any stage and respond with a proper message and HTTP status code. For example, consider the scenario in which you are trying to create a resource with a referential integrity constraint at the db schema. We can set the emco error message with a simple, meaningful message, for example, failed to create the resource. We can use the cause parameter to set the actual/original error, in this case, the referential integrity requirement, thus providing more information to the user and tracing the errors across the stack. 

``` sh
    func (c *CaCertClient) CreateCert(cert CaCert, failIfExists bool) (CaCert, bool, error) {
        if err := db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, cert); err != nil {
            return CaCert{}, certExists, &emcoerror.Error{
                Message: "Failed to create the caCert",
                Reason:  err.Reason,
                Cause: err,
            }
	    }
    }

    func (m *MongoStore) Insert(coll string, key Key, query interface{}, tag string, data interface{}) error {
        return &emcoerror.Error{
            Message: fmt.Sprintf("Parent resource not found for %s.  Parent: %T %v KeyID: %s, Key: %T %v", name, parentKey, parentKey, keyId, key, key),
            Reason:  emcoerror.Conflict, // In this case, the reason can be a db-specific reason, like invalidSchema, which can be mapped with an HTTP status code.
        }
    }
```


### Handling life cycle errors

Sometimes the life-cycle operations can fail due to numerous reasons. These errors can be handled and managed using the emcoerror package. 
The emcoerror package defines a `stateerror` type. 

``` sh
    type StateError struct {
        Resource string 
        Event    string 
        Status   appcontext.StatusValue
    }

    func (e *StateError) Error() string {
    }
```

The emco `stateerror` type consists of the following.

- `Resource` - This is the name of the resource which supports the life-cycle events. e.g: LogicalCloud, DeploymentIntentGroup, CaCert etc.
- `Event` - This is the life-cycle event. e.g: Instantiate, Terminate etc.
- `Status` - This is the current status of the resource

The emcoerror package handles the life-cycle failures and generates the appropriate error message to send back to the user. For example, terminating the caCert distribution is a life-cycle operation. Consider the scenario,where the caCert distribution is in a terminating status, and the user is trying to terminate it again.  In this case, this is a conflict. The emcoerror package handles this as follows.

``` sh
    func (sc *StateClient) VerifyState(event LifeCycleEvent) (string, error) {
        switch status.Status {
        case appcontext.AppContextStatusEnum.Terminating:
            err := &emcoerror.Error{
                Message: (&emcoerror.StateError{
                    Resource: "CaCert",
                    Event:    string(event),
                    Status:   status.Status, 
                    }).Error(),
                Reason: emcoerror.Conflict,
                }
            logutils.Error("",
                logutils.Fields{
                    "Error":     err.Error(),
                    "ContextID": contextID})
            return contextID, err
        }
    }


    func (e *StateError) Error() string {
        switch e.Status {
        case appcontext.AppContextStatusEnum.Terminating:
            return fmt.Sprintf("Failed to %s. The %s is being terminated", e.Event, e.Resource)
        }
    }
```

First, we create a `stateerror` type, which implements the error interface to generate the appropriate message that needs to be logged or returned to the user. In this example, the stateerror will return the message `Failed to terminate. The CaCert is being terminated`. Now the emco error type defines the reason for this error as a conflict. The module returns the emco error to the handler to format the response and status code based on the error reason.

The controller should decide whether an error needs to be returned or not during a life-cycle event. For example, we should return an error if the user tries to `terminate` a resource in the `terminated` status. The user should be able to continue with `instantiation` even if it is in the `terminated` status. The emcoerror package does not make these decisions based on the event or resource status. We should use the emcoerror package only to format the error message. The caCert controller verifies the state of the enrollment/ distribution of the caCert intent as follows.

``` sh
func (sc *StateClient) VerifyState(event common.EmcoEvent) (string, error) {
	switch status.Status {
		case appcontext.AppContextStatusEnum.Terminated: `The resource is in the terminated status`.
			// handle events specific use cases
			switch event {
			case common.Instantiate: `Continue with the instantiation.`
				return contextID, nil  
			case common.Terminate: `The resource is already terminated. Return an error to the user.`
				err := &emcoerror.Error{
					Message: (&emcoerror.StateError{ `Use the emco error package to format the error message`
						Resource: "CaCert",
						Event:    event,
						Status:   status.Status}).Error(),
					Reason: emcoerror.Conflict,
				}
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			}
    }
}
```

### Managing anticipated errors

Since each emco error type is associated with a specific reason, we can explicitly verify them anywhere in the stack. Sometimes we need to make logical decisions based on the kind of failure we encounter during the execution stack. For example, we need to verify the caCert enrollment state before deleting it. Some validation errors will stop you from the caCert deletion, but some may not. The emco error reason helps in managing these kinds of anticipated errors. Using the error reason also avoids the overhead of string/ error comparison.

``` sh
    func (c *CaCertClient) DeleteCert(cert, clusterProvider string) error {
        // check the enrollment state
        if err := verifyEnrollmentStateBeforeDelete(cert, clusterProvider); err != nil {
            // if the StateInfo cannot be found, then a caCert record may not present
            // Continue with the caCert deletion if the error is NotFound
            // In all other cases, intercept and return the error
            switch e := err.(type) {
            case *emcoerror.Error:
                if e.Reason != emcoerror.NotFound {
                    return e
                }
            default:
                return err
            }
        }
    }
```

> **NOTE**: The type switch, in this case, it avoids any panic because the error received from the calling stack can be an emco error, the built-in error, or any other custom-defined error. 


## Defining error messages

Another aspect of error handling is the messages. We should always respond with consistent and meaningful error messages. We can define the possible expected error messages in which a package or controller can use to set the emco error message. It will also avoid duplicating the string at different places in a different style ( casing, format, etc.). For example, the caCert controller module package defines some of the commonly used error messages at the module level. Other packages in the caCert controller, like the clusterprovider package, can use these defined errors to set the emco error message as needed. We can use this approach to create the emco error anywhere in the code. 

``` sh
    package module

    // caCert errors
    const (
        CaCertAlreadyExists  string = "caCert already exists"
    )

    package clusterprovider

    func (c *CaCertClient) CreateCert(cert module.CaCert, clusterProvider string, failIfExists bool) (module.CaCert, bool, error) {
        if certExists &&
            failIfExists {
            return module.CaCert{}, certExists, &emcoerror.Error{
                Message: module.CaCertAlreadyExists,
                Reason:  emcoerror.Conflict,
            }
        }
    }
```
