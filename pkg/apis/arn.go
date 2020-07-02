/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apis

import (
	"encoding"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"
)

// ARN extends arn.ARN with additional methods for (de-)serialization to/from
// JSON, allowing it to be embedded in custom API objects.
type ARN arn.ARN

var (
	_ fmt.Stringer             = (*ARN)(nil)
	_ encoding.TextMarshaler   = (*ARN)(nil)
	_ encoding.TextUnmarshaler = (*ARN)(nil)
)

// String implements the fmt.Stringer interface.
func (a ARN) String() string {
	return arn.ARN(a).String()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (a *ARN) UnmarshalText(data []byte) error {
	dataStr := string(data)
	arn, err := arn.Parse(dataStr)
	if err != nil {
		return fmt.Errorf("failed to parse ARN %q: %w", dataStr, err)
	}

	*a = ARN(arn)

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (a ARN) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}
