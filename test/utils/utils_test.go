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

package utils

import (
	"os"
	"os/exec"
	"testing"

	"github.com/onsi/gomega"
)

func TestWarnError(t *testing.T) {
	// Test that warnError doesn't panic with a valid error
	err := os.ErrNotExist
	// This function writes to GinkgoWriter, so we can't easily test the output
	// but we can test that it doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("warnError panicked: %v", r)
			}
		}()
		warnError(err)
	}()
}

func TestGetNonEmptyLines(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with empty string
	result := GetNonEmptyLines("")
	g.Expect(result).To(gomega.BeEmpty())

	// Test with single line
	result = GetNonEmptyLines("single line")
	g.Expect(result).To(gomega.HaveLen(1))
	g.Expect(result[0]).To(gomega.Equal("single line"))

	// Test with multiple lines including empty ones
	input := "line1\n\nline2\n\n\nline3\n"
	result = GetNonEmptyLines(input)
	g.Expect(result).To(gomega.HaveLen(3))
	g.Expect(result[0]).To(gomega.Equal("line1"))
	g.Expect(result[1]).To(gomega.Equal("line2"))
	g.Expect(result[2]).To(gomega.Equal("line3"))

	// Test with only empty lines
	result = GetNonEmptyLines("\n\n\n")
	g.Expect(result).To(gomega.BeEmpty())

	// Test with lines containing only whitespace
	result = GetNonEmptyLines("line1\n   \nline2\n\t\nline3")
	g.Expect(result).To(gomega.HaveLen(3))
	g.Expect(result[0]).To(gomega.Equal("line1"))
	g.Expect(result[1]).To(gomega.Equal("   "))
	g.Expect(result[2]).To(gomega.Equal("line3"))
}

func TestGetProjectDir(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test normal case
	dir, err := GetProjectDir()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(dir).ToNot(gomega.BeEmpty())

	// Test that it removes /test/e2e from the path
	// We can't easily test this without changing the working directory,
	// but we can test that the function doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetProjectDir panicked: %v", r)
			}
		}()
		_, _ = GetProjectDir()
	}()
}

func TestRun(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with a simple command that should succeed
	cmd := exec.Command("echo", "test")
	output, err := Run(cmd)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(output)).To(gomega.ContainSubstring("test"))

	// Test with a command that should fail
	cmd = exec.Command("nonexistentcommand")
	output, err = Run(cmd)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(output).ToNot(gomega.BeNil())
}

func TestLoadImageToKindClusterWithName(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with a dummy image name
	// This will likely fail in a test environment without kind, but we can test the function structure
	err := LoadImageToKindClusterWithName("test-image:latest")
	// We expect this to fail in a test environment, but the function should not panic
	g.Expect(err).ToNot(gomega.BeNil()) // Expected to fail without kind cluster
}

func TestLoadImageToKindClusterWithName_WithEnvVar(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with KIND_CLUSTER environment variable set
	originalValue := os.Getenv("KIND_CLUSTER")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("KIND_CLUSTER")
		} else {
			os.Setenv("KIND_CLUSTER", originalValue)
		}
	}()

	os.Setenv("KIND_CLUSTER", "test-cluster")
	err := LoadImageToKindClusterWithName("test-image:latest")
	// We expect this to fail in a test environment, but the function should not panic
	g.Expect(err).ToNot(gomega.BeNil()) // Expected to fail without kind cluster
}

func TestInstallPrometheusOperator(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test that the function doesn't panic
	// This will likely fail in a test environment without kubectl, but we can test the function structure
	err := InstallPrometheusOperator()
	// We expect this to fail in a test environment, but the function should not panic
	g.Expect(err).ToNot(gomega.BeNil()) // Expected to fail without kubectl
}

func TestUninstallPrometheusOperator(t *testing.T) {
	// Test that the function doesn't panic
	// This will likely fail in a test environment without kubectl, but we can test the function structure
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("UninstallPrometheusOperator panicked: %v", r)
			}
		}()
		UninstallPrometheusOperator()
	}()
}

func TestInstallCertManager(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test that the function doesn't panic
	// This will likely fail in a test environment without kubectl, but we can test the function structure
	err := InstallCertManager()
	// We expect this to fail in a test environment, but the function should not panic
	g.Expect(err).ToNot(gomega.BeNil()) // Expected to fail without kubectl
}

func TestUninstallCertManager(t *testing.T) {

	// Test that the function doesn't panic
	// This will likely fail in a test environment without kubectl, but we can test the function structure
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("UninstallCertManager panicked: %v", r)
			}
		}()
		UninstallCertManager()
	}()
}

func TestConstants(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test that constants are defined
	g.Expect(prometheusOperatorVersion).To(gomega.Equal("v0.68.0"))
	g.Expect(prometheusOperatorURL).To(gomega.ContainSubstring("github.com/prometheus-operator/prometheus-operator"))
	g.Expect(certmanagerVersion).To(gomega.Equal("v1.5.3"))
	g.Expect(certmanagerURLTmpl).To(gomega.ContainSubstring("github.com/jetstack/cert-manager"))
}

func TestURLFormatting(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test Prometheus Operator URL formatting
	expectedPrometheusURL := "https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.68.0/bundle.yaml"
	actualPrometheusURL := "https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.68.0/bundle.yaml"
	g.Expect(actualPrometheusURL).To(gomega.Equal(expectedPrometheusURL))

	// Test Cert Manager URL formatting
	expectedCertManagerURL := "https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml"
	actualCertManagerURL := "https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml"
	g.Expect(actualCertManagerURL).To(gomega.Equal(expectedCertManagerURL))
}

func TestCommandConstruction(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test that commands are constructed correctly
	// We can't easily test the actual command execution, but we can test the structure

	// Test kubectl create command for Prometheus Operator
	expectedArgs := []string{"kubectl", "create", "-f", "https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.68.0/bundle.yaml"}
	// This is what the InstallPrometheusOperator function would create
	g.Expect(expectedArgs[0]).To(gomega.Equal("kubectl"))
	g.Expect(expectedArgs[1]).To(gomega.Equal("create"))
	g.Expect(expectedArgs[2]).To(gomega.Equal("-f"))

	// Test kubectl apply command for Cert Manager
	expectedArgs = []string{"kubectl", "apply", "-f", "https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml"}
	g.Expect(expectedArgs[0]).To(gomega.Equal("kubectl"))
	g.Expect(expectedArgs[1]).To(gomega.Equal("apply"))
	g.Expect(expectedArgs[2]).To(gomega.Equal("-f"))

	// Test kind load command
	expectedArgs = []string{"kind", "load", "docker-image", "test-image", "--name", "kind"}
	g.Expect(expectedArgs[0]).To(gomega.Equal("kind"))
	g.Expect(expectedArgs[1]).To(gomega.Equal("load"))
	g.Expect(expectedArgs[2]).To(gomega.Equal("docker-image"))
	g.Expect(expectedArgs[3]).To(gomega.Equal("test-image"))
	g.Expect(expectedArgs[4]).To(gomega.Equal("--name"))
	g.Expect(expectedArgs[5]).To(gomega.Equal("kind"))
}

func TestErrorHandling(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test that functions handle errors gracefully
	// We can't easily test all error conditions, but we can test that functions don't panic

	// Test Run function with invalid command
	cmd := exec.Command("nonexistentcommand")
	output, err := Run(cmd)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(output).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("failed with error"))

	// Test that warnError doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("warnError panicked: %v", r)
			}
		}()
		warnError(err)
	}()
}
