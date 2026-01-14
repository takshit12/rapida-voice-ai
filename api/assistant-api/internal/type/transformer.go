// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_type

import (
	"context"
)

// Transformers is a generic interface that defines a transform method
// for converting one type to another.
//
// Type parameters:
//   - IN: The input type for the transformation.
//   - OUT: The output type for the transformation.
//
// The transform method takes an input of type IN and returns an output
// of type OUT along with an error. This allows for flexible type
// conversion and data transformation while providing error handling.
//
// Implementations of this interface can be used to create reusable
// and composable transformation logic for various data types and
// structures within an application.

type Transformers[IN any] interface {
	// The `Initialize() error` method in the `Transformers` interface is defining a function signature
	// for a method that initializes or sets up any necessary resources or configurations before the
	// transformation process begins. This method is expected to return an error if any issues occur
	// during the initialization process, allowing for proper error handling and ensuring that the
	// transformation can proceed only when the initialization is successful.
	Initialize() error

	// The comment `// Transformers[[]byte, string]` is specifying the type parameters for the interface
	// `SpeechToTextTransformer`. It is indicating that `SpeechToTextTransformer` extends the `Transformers`
	// interface with the specific type parameters `[]byte` as the input type and `string` as the output
	// type for the transformation.
	Transform(context.Context, IN) error

	//
	// The `Cancel() error` method in the `Transformers` interface defines a function signature for a
	// method that aborts or cancels an ongoing transformation process. This method is expected to
	// return an error if any issues occur during the cancellation process.
	Close(context.Context) error
}
