/*
GNU GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (c) 2024-25 Norihiro Seto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

For the full license text, please visit: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package controller

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	specAnnotation = "LBRegistrar.nodes.peppy-ratio.dev/spec"
)

func hasSpecAnnotation(resource *metav1.ObjectMeta) bool {
	return resource.Annotations != nil && resource.Annotations[specAnnotation] != ""
}

func saveSpecInAnnotations(resource *metav1.ObjectMeta, spec any) error {
	if !hasSpecAnnotation(resource) {
		resource.Annotations = make(map[string]string)
	}
	serializedSpec, err := serializeSpec(spec)
	if err != nil {
		return errors.NewInternalError(err)
	}
	resource.Annotations[specAnnotation] = serializedSpec

	return nil
}

func getSpecInAnnotations(resource *metav1.ObjectMeta, spec any) (bool, error) {
	if !hasSpecAnnotation(resource) {
		return false, nil
	}
	serializedSpec, ok := resource.Annotations[specAnnotation]
	if !ok {
		return false, nil
	}
	if err := deserializeSpec(serializedSpec, spec); err != nil {
		return false, errors.NewInternalError(err)
	}
	return true, nil
}

func deleteSpecInAnnotations(resource *metav1.ObjectMeta) {
	if !hasSpecAnnotation(resource) {
		return
	}
	delete(resource.Annotations, specAnnotation)
}

func serializeSpec(spec any) (string, error) {
	bytes, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func deserializeSpec(serializedSpec string, spec any) error {
	return json.Unmarshal([]byte(serializedSpec), spec)
}
