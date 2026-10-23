/*
Copyright 2026 The cert-manager Authors.

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

package certificates

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/utils/clock"

	"github.com/cert-manager/cert-manager/integration-tests/framework"
	"github.com/cert-manager/cert-manager/internal/controller/feature"
	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	controllerpkg "github.com/cert-manager/cert-manager/pkg/controller"
	"github.com/cert-manager/cert-manager/pkg/controller/certificates/keymanager"
	logf "github.com/cert-manager/cert-manager/pkg/logs"
	"github.com/cert-manager/cert-manager/pkg/metrics"
	utilfeature "github.com/cert-manager/cert-manager/pkg/util/feature"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
)

// TestDefaultPrivateKeyRotationPolicy verifies that the controller actually
// reuses or replaces an existing key according to the gate and explicit policy.
func TestDefaultPrivateKeyRotationPolicy(t *testing.T) {
	for _, test := range []struct {
		name    string
		enabled bool
		policy  cmapi.PrivateKeyRotationPolicy
		reuse   bool
	}{
		{name: "enabled-default-always", enabled: true},
		{name: "disabled-default-never", enabled: false, reuse: true},
		{name: "disabled-explicit-always", enabled: false, policy: cmapi.RotationPolicyAlways},
		{name: "enabled-explicit-never", enabled: true, policy: cmapi.RotationPolicyNever, reuse: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, feature.DefaultPrivateKeyRotationPolicyAlways, test.enabled)
			config, stopPlane := framework.RunControlPlane(t)
			t.Cleanup(stopPlane)
			kubeClient, factory, cmClient, cmFactory, scheme := framework.NewClients(t, config)

			crt := &cmapi.Certificate{
				ObjectMeta: metav1.ObjectMeta{Name: "rotation-policy", Namespace: "default"},
				Spec: cmapi.CertificateSpec{
					SecretName: "existing-key",
					DNSNames:   []string{"example.test"},
					IssuerRef:  cmmeta.IssuerReference{Name: "test-issuer"},
					PrivateKey: &cmapi.CertificatePrivateKey{
						Algorithm:      cmapi.ECDSAKeyAlgorithm,
						Size:           256,
						RotationPolicy: test.policy,
					},
				},
			}
			originalKey, err := pki.GeneratePrivateKeyForCertificate(crt)
			require.NoError(t, err)
			originalPEM, err := pki.EncodePrivateKey(originalKey, cmapi.PKCS8)
			require.NoError(t, err)
			_, err = kubeClient.CoreV1().Secrets(crt.Namespace).Create(t.Context(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: crt.Spec.SecretName, Namespace: crt.Namespace},
				Data:       map[string][]byte{corev1.TLSPrivateKeyKey: originalPEM},
			}, metav1.CreateOptions{})
			require.NoError(t, err)
			crt, err = cmClient.CertmanagerV1().Certificates(crt.Namespace).Create(t.Context(), crt, metav1.CreateOptions{})
			require.NoError(t, err)
			apiutil.SetCertificateCondition(crt, crt.Generation, cmapi.CertificateConditionIssuing, cmmeta.ConditionTrue, "TestIssuance", "test private key rotation")
			_, err = cmClient.CertmanagerV1().Certificates(crt.Namespace).UpdateStatus(t.Context(), crt, metav1.UpdateOptions{})
			require.NoError(t, err)

			metric := metrics.New(logf.Log, clock.RealClock{})
			controllerContext := &controllerpkg.Context{
				Client:                    kubeClient,
				Scheme:                    scheme,
				KubeSharedInformerFactory: factory,
				CMClient:                  cmClient,
				SharedInformerFactory:     cmFactory,
				Metrics:                   metric,
				Clock:                     clock.RealClock{},
				Recorder:                  framework.NewEventRecorder(t, scheme),
				FieldManager:              "rotation-policy-test",
			}
			ctrl, queue, mustSync, err := keymanager.NewController(logf.Log, controllerContext)
			require.NoError(t, err)
			manager := controllerpkg.NewController("rotation-policy-test", metric, ctrl.ProcessItem, mustSync, nil, queue)
			stopController := framework.StartInformersAndController(t, factory, cmFactory, manager)
			defer stopController()

			var nextSecret *corev1.Secret
			err = wait.PollUntilContextTimeout(t.Context(), 100*time.Millisecond, 20*time.Second, true, func(ctx context.Context) (bool, error) {
				stored, err := cmClient.CertmanagerV1().Certificates(crt.Namespace).Get(ctx, crt.Name, metav1.GetOptions{})
				if err != nil || stored.Status.NextPrivateKeySecretName == nil {
					return false, err
				}
				require.Equal(t, crt.Spec.PrivateKey, stored.Spec.PrivateKey, "runtime defaulting must not change the stored Certificate spec")
				nextSecret, err = kubeClient.CoreV1().Secrets(crt.Namespace).Get(ctx, *stored.Status.NextPrivateKeySecretName, metav1.GetOptions{})
				return err == nil, err
			})
			require.NoError(t, err)
			nextKey, err := pki.DecodePrivateKeyBytes(nextSecret.Data[corev1.TLSPrivateKeyKey])
			require.NoError(t, err)
			sameKey, err := pki.PublicKeysEqual(originalKey.Public(), nextKey.Public())
			require.NoError(t, err)
			require.Equal(t, test.reuse, sameKey, "private key reuse must follow the effective rotation policy")
		})
	}
}
