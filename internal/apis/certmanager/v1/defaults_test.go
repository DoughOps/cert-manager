/*
Copyright 2020 The cert-manager Authors.

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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	"github.com/cert-manager/cert-manager/internal/controller/feature"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	utilfeature "github.com/cert-manager/cert-manager/pkg/util/feature"
)

// Test_SetRuntimeDefaults_Certificate_PrivateKey_RotationPolicy demonstrates that
// the default rotation policy follows the feature gate without overriding an
// explicitly configured policy.
func Test_SetRuntimeDefaults_Certificate_PrivateKey_RotationPolicy(t *testing.T) {
	t.Run("default-rotation-policy", func(t *testing.T) {
		require.True(t, utilfeature.DefaultFeatureGate.Enabled(feature.DefaultPrivateKeyRotationPolicyAlways))
		in := &cmapi.Certificate{}
		SetRuntimeDefaults_Certificate(in)
		assert.Equal(t, cmapi.RotationPolicyAlways, in.Spec.PrivateKey.RotationPolicy)
	})
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "feature-enabled", false: "feature-disabled"}[enabled], func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, feature.DefaultPrivateKeyRotationPolicyAlways, enabled)
			defaultPolicy := cmapi.RotationPolicyNever
			if enabled {
				defaultPolicy = cmapi.RotationPolicyAlways
			}
			for _, test := range []struct {
				name       string
				privateKey *cmapi.CertificatePrivateKey
			}{
				{name: "omitted-private-key"},
				{name: "empty-private-key", privateKey: &cmapi.CertificatePrivateKey{}},
				{name: "key-options", privateKey: &cmapi.CertificatePrivateKey{Algorithm: cmapi.ECDSAKeyAlgorithm, Size: 256, Encoding: cmapi.PKCS8}},
				{name: "explicit-never", privateKey: &cmapi.CertificatePrivateKey{RotationPolicy: cmapi.RotationPolicyNever}},
				{name: "explicit-always", privateKey: &cmapi.CertificatePrivateKey{RotationPolicy: cmapi.RotationPolicyAlways}},
				{name: "unknown-policy", privateKey: &cmapi.CertificatePrivateKey{RotationPolicy: "neither-always-nor-never"}},
			} {
				t.Run(test.name, func(t *testing.T) {
					in := &cmapi.Certificate{}
					if test.privateKey != nil {
						in.Spec.PrivateKey = test.privateKey.DeepCopy()
					}
					expected := &cmapi.CertificatePrivateKey{}
					if test.privateKey != nil {
						expected = test.privateKey.DeepCopy()
					}
					if expected.RotationPolicy == "" {
						expected.RotationPolicy = defaultPolicy
					}
					SetRuntimeDefaults_Certificate(in)
					assert.Equal(t, expected, in.Spec.PrivateKey)
					SetRuntimeDefaults_Certificate(in)
					assert.Equal(t, expected, in.Spec.PrivateKey, "defaulting must be idempotent")
				})
			}
		})
	}
}
