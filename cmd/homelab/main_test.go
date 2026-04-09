package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateServiceFile(t *testing.T) {
	name := "testApp"

	serviceContent := generateServiceFile(name)

	// Verificações principais (não testa tudo, só o importante)
	if !strings.Contains(serviceContent, "Description=testApp") {
		t.Errorf("expected Description to contain app name")
	}

	if !strings.Contains(serviceContent, "ExecStart=/opt/homelab/testApp") {
		t.Errorf("expected ExecStart to be correct")
	}

	if !strings.Contains(serviceContent, "Restart=always") {
		t.Errorf("expected Restart=always")
	}
	fmt.Println(serviceContent)
}
