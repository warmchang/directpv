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

package types

import (
	"fmt"
	"strings"

	"github.com/minio/directpv/pkg/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
)

// LabelKey stores label keys
type LabelKey string

const (
	// PodNameLabelKey label key for pod name
	PodNameLabelKey LabelKey = consts.GroupName + "/pod.name"

	// PodNSLabelKey label key for pod namespace
	PodNSLabelKey LabelKey = consts.GroupName + "/pod.namespace"

	// NodeLabelKey label key for node
	NodeLabelKey LabelKey = consts.GroupName + "/node"

	// DriveLabelKey label key for drive
	DriveLabelKey LabelKey = consts.GroupName + "/drive"

	// PathLabelKey key for path
	PathLabelKey LabelKey = consts.GroupName + "/path"

	// AccessTierLabelKey label key for access-tier
	AccessTierLabelKey LabelKey = consts.GroupName + "/access-tier"

	// VersionLabelKey label key for version
	VersionLabelKey LabelKey = consts.GroupName + "/version"

	// CreatedByLabelKey label key for created by
	CreatedByLabelKey LabelKey = consts.GroupName + "/created-by"

	// DrivePathLabelKey label key for drive path
	DrivePathLabelKey LabelKey = consts.GroupName + "/drive-path"

	// LatestVersionLabelKey label key for group and version
	LatestVersionLabelKey LabelKey = consts.GroupName + "/" + consts.LatestAPIVersion

	// TopologyDriverIdentity label key for identity
	TopologyDriverIdentity LabelKey = consts.GroupName + "/identity"

	// TopologyDriverNode label key for node
	TopologyDriverNode LabelKey = consts.GroupName + "/node"

	// TopologyDriverRack label key for rack
	TopologyDriverRack LabelKey = consts.GroupName + "/rack"

	// TopologyDriverZone label key for zone
	TopologyDriverZone LabelKey = consts.GroupName + "/zone"

	// TopologyDriverRegion label key for region
	TopologyDriverRegion LabelKey = consts.GroupName + "/region"
)

// LabelValue is a type definition for label value
type LabelValue string

// NewLabelValue validates and converts string value to label value
func NewLabelValue(value string) LabelValue {
	errs := validation.IsValidLabelValue(value)
	if len(errs) == 0 {
		return LabelValue(value)
	}

	normalizeLabelValue := func(value string) string {
		if len(value) > 63 {
			value = value[:63]
		}

		result := []rune(value)
		for i, r := range result {
			switch {
			case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			default:
				if i != 0 && r != '.' && r != '_' && r != '-' {
					result[i] = '-'
				} else {
					result[i] = 'x'
				}
			}
		}

		return string(result)
	}

	result := LabelValue(normalizeLabelValue(value))
	klog.V(3).InfoS(
		fmt.Sprintf("label value converted due to invalid value; %v", strings.Join(errs, "; ")),
		"value", value, "converted value", result,
	)
	return result
}

// ToLabelSelector converts a map of label key and label value to selector string
func ToLabelSelector(labels map[LabelKey][]LabelValue) string {
	selectors := []string{}
	for key, values := range labels {
		if len(values) != 0 {
			result := []string{}
			for _, value := range values {
				result = append(result, string(value))
			}
			selectors = append(selectors, fmt.Sprintf("%s in (%s)", key, strings.Join(result, ",")))
		}
	}
	return strings.Join(selectors, ",")
}

// SetLabels sets labels in object.
func SetLabels(object metav1.Object, labels map[LabelKey]LabelValue) {
	values := make(map[string]string)
	for key, value := range labels {
		values[string(key)] = string(value)
	}
	object.SetLabels(values)
}

// UpdateLabels updates labels in object.
func UpdateLabels(object metav1.Object, labels map[LabelKey]LabelValue) {
	values := object.GetLabels()
	if values == nil {
		values = make(map[string]string)
	}

	for key, value := range labels {
		values[string(key)] = string(value)
	}

	object.SetLabels(values)
}