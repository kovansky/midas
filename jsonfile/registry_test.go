/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package jsonfile

import (
	"errors"
	"github.com/kovansky/midas"
	"os"
	"testing"
)

var (
	site = midas.Site{RootDir: "./", Registry: midas.RegistrySettings{Type: "jsonfile", Location: "./test-registry.json"}}
	r    = NewRegistryService(site).(*RegistryService)
)

func TestRegistryService_OpenStorage(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Open", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.OpenStorage(); (err != nil) != tt.wantErr {
				t.Errorf("OpenStorage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, err := os.Stat(r.path); errors.Is(err, os.ErrNotExist) {
				t.Errorf("OpenStorage() registry file wasn't created")
			}
		})
	}
}

func TestRegistryService_CreateEntry(t *testing.T) {
	type args struct {
		id       string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{"First", args{"test-1", "test-1.html"}, 1, false},
		{"Second", args{"test-2", "test-2.html"}, 2, false},
		{"Duplicate", args{"test-2", "test-3.html"}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.CreateEntry(tt.args.id, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("CreateEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(r.registry) != tt.wantLen {
				t.Errorf("CreateEntry() len = %v, wantLen %v", len(r.registry), tt.wantLen)
			}
		})
	}
}

func TestRegistryService_ReadEntry(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"FirstExisting", args{"test-1"}, "test-1.html", false},
		{"SecondExisting", args{"test-2"}, "test-2.html", false},
		{"Nonexisting", args{"test-3"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.ReadEntry(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadEntry() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistryService_UpdateEntry(t *testing.T) {
	type args struct {
		id          string
		newFilename string
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{"Existing", args{"test-2", "test-other.html"}, 2, false},
		{"Nonexisting", args{"test-3", "foobar.html"}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.UpdateEntry(tt.args.id, tt.args.newFilename); (err != nil) != tt.wantErr {
				t.Errorf("UpdateEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(r.registry) != tt.wantLen {
				t.Errorf("UpdateEntry() len = %v, wantLen %v", len(r.registry), tt.wantLen)
			}
		})
	}
}

func TestRegistryService_DeleteEntry(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{"Existing", args{"test-2"}, 1, false},
		{"Nonexisting", args{"test-3"}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.DeleteEntry(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(r.registry) != tt.wantLen {
				t.Errorf("DeleteEntry() len = %v, wantLen %v", len(r.registry), tt.wantLen)
			}
		})
	}
}

func TestRegistryService_Flush(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Flush", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lenBefore := len(r.registry)

			if err := r.Flush(); (err != nil) != tt.wantErr {
				t.Errorf("Flush() error = %v, wantErr %v", err, tt.wantErr)
			}

			err := r.readStorage()
			if err != nil {
				t.Errorf("readStorage() error = %v", err)
			}

			if len(r.registry) != lenBefore {
				t.Errorf("Flush() len before flush = %v, after flush = %v", lenBefore, len(r.registry))
			}
		})
	}
}

func TestRegistryService_RemoveStorage(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Remove", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.RemoveStorage(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveStorage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, err := os.Stat(r.path); err == nil {
				t.Errorf("RemoveStorage() registry file wasn't removed")
			}
		})
	}
}
