// This file is part of MinIO DirectPV
// Copyright (c) 2021, 2022 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package k8s

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/minio/directpv/pkg/utils"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// MaxThreadCount is maximum thread count.
const MaxThreadCount = 200

// GetKubeConfig gets kubernetes configuration from command line argument,
// ~/.kube/config or in-cluster configuration.
func GetKubeConfig() (*rest.Config, error) {
	kubeconfigPath := viper.GetString("kubeconfig")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			klog.ErrorS(err, "unable to find user home directory")
		}
		kubeconfigPath = path.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	}

	config.QPS = float32(MaxThreadCount / 2)
	config.Burst = MaxThreadCount
	return config, nil
}

// IsCondition checks whether type/status/reason/message in conditions or not.
func IsCondition(conditions []metav1.Condition, ctype string, status metav1.ConditionStatus, reason, message string) bool {
	for i := range conditions {
		if conditions[i].Type == ctype &&
			conditions[i].Status == status &&
			conditions[i].Reason == reason &&
			conditions[i].Message == message {
			return true
		}
	}
	return false
}

// UpdateCondition updates type/status/reason/message of conditions matched by condition type.
func UpdateCondition(conditions []metav1.Condition, ctype string, status metav1.ConditionStatus, reason, message string) {
	for i := range conditions {
		if conditions[i].Type == ctype {
			conditions[i].Status = status
			conditions[i].Reason = reason
			conditions[i].Message = message
			conditions[i].LastTransitionTime = metav1.Now()
			break
		}
	}
}

// MatchTrueConditions matches whether types and status list are in a true conditions or not.
func MatchTrueConditions(conditions []metav1.Condition, types, statusList []string) bool {
	for i := range types {
		types[i] = strings.ToLower(types[i])
	}
	for i := range statusList {
		statusList[i] = strings.ToLower(statusList[i])
	}

	statusMatches := 0
	for _, condition := range conditions {
		ctype := strings.ToLower(condition.Type)
		if condition.Status == metav1.ConditionTrue && utils.Contains(types, ctype) && utils.Contains(statusList, ctype) {
			statusMatches++
		}
	}

	return statusMatches == len(statusList)
}

// BoolToConditionStatus converts boolean value to condition status.
func BoolToConditionStatus(val bool) metav1.ConditionStatus {
	if val {
		return metav1.ConditionTrue
	}
	return metav1.ConditionFalse
}

// SanitizeResourceName - Sanitize given name to a valid kubernetes name format.
// RegEx for a kubernetes name is
//
//	([a-z0-9][-a-z0-9]*)?[a-z0-9]
//
// with a max length of 253
//
// WARNING: This function will truncate to 253 bytes if the input is longer
func SanitizeResourceName(name string) string {
	if len(name) > 253 {
		name = name[:253]
	}

	result := []rune(strings.ToLower(name))
	for i, r := range result {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
		default:
			if i == 0 {
				result[i] = '0'
			} else {
				result[i] = '-'
			}
		}
	}

	return string(result)
}

// ParseNodeSelector parses the provided node selector values
func ParseNodeSelector(values []string) (map[string]string, error) {
	nodeSelector := map[string]string{}
	for _, value := range values {
		tokens := strings.Split(value, "=")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("invalid node selector value %v", value)
		}
		if tokens[0] == "" {
			return nil, fmt.Errorf("invalid key in node selector value %v", value)
		}
		nodeSelector[tokens[0]] = tokens[1]
	}
	return nodeSelector, nil
}

// ParseTolerations parses the provided toleration values
func ParseTolerations(values []string) ([]corev1.Toleration, error) {
	var tolerations []corev1.Toleration
	for _, value := range values {
		var k, v, e string
		tokens := strings.SplitN(value, "=", 2)
		switch len(tokens) {
		case 1:
			k = tokens[0]
			tokens = strings.Split(k, ":")
			switch len(tokens) {
			case 1:
			case 2:
				k, e = tokens[0], tokens[1]
			default:
				if len(tokens) != 2 {
					return nil, fmt.Errorf("invalid toleration %v", value)
				}
			}
		case 2:
			k, v = tokens[0], tokens[1]
		default:
			if len(tokens) != 2 {
				return nil, fmt.Errorf("invalid toleration %v", value)
			}
		}
		if k == "" {
			return nil, fmt.Errorf("invalid key in toleration %v", value)
		}
		if v != "" {
			if tokens = strings.Split(v, ":"); len(tokens) != 2 {
				return nil, fmt.Errorf("invalid value in toleration %v", value)
			}
			v, e = tokens[0], tokens[1]
		}
		effect := corev1.TaintEffect(e)
		switch effect {
		case corev1.TaintEffectNoSchedule, corev1.TaintEffectPreferNoSchedule, corev1.TaintEffectNoExecute:
		default:
			return nil, fmt.Errorf("invalid toleration effect in toleration %v", value)
		}
		operator := corev1.TolerationOpExists
		if v != "" {
			operator = corev1.TolerationOpEqual
		}
		tolerations = append(tolerations, corev1.Toleration{
			Key:      k,
			Operator: operator,
			Value:    v,
			Effect:   effect,
		})
	}

	return tolerations, nil
}

// NewHostPathVolume - creates volume for given name and host path.
func NewHostPathVolume(name, path string) corev1.Volume {
	hostPathType := corev1.HostPathDirectoryOrCreate
	volumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
			Type: &hostPathType,
		},
	}

	return corev1.Volume{
		Name:         name,
		VolumeSource: volumeSource,
	}
}

// NewVolumeMount - creates volume mount for given name, path, mount propagation and read only flag.
func NewVolumeMount(name, path string, mountPropogation corev1.MountPropagationMode, readOnly bool) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:             name,
		ReadOnly:         readOnly,
		MountPath:        path,
		MountPropagation: &mountPropogation,
	}
}
