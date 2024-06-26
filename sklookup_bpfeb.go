// Code generated by bpf2go; DO NOT EDIT.
//go:build mips || mips64 || ppc64 || s390x

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

// loadSklookup returns the embedded CollectionSpec for sklookup.
func loadSklookup() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_SklookupBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load sklookup: %w", err)
	}

	return spec, err
}

// loadSklookupObjects loads sklookup and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*sklookupObjects
//	*sklookupPrograms
//	*sklookupMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadSklookupObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadSklookup()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// sklookupSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type sklookupSpecs struct {
	sklookupProgramSpecs
	sklookupMapSpecs
}

// sklookupSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type sklookupProgramSpecs struct {
	EchoDispatch *ebpf.ProgramSpec `ebpf:"echo_dispatch"`
}

// sklookupMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type sklookupMapSpecs struct {
	EchoPorts  *ebpf.MapSpec `ebpf:"echo_ports"`
	EchoSocket *ebpf.MapSpec `ebpf:"echo_socket"`
}

// sklookupObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadSklookupObjects or ebpf.CollectionSpec.LoadAndAssign.
type sklookupObjects struct {
	sklookupPrograms
	sklookupMaps
}

func (o *sklookupObjects) Close() error {
	return _SklookupClose(
		&o.sklookupPrograms,
		&o.sklookupMaps,
	)
}

// sklookupMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadSklookupObjects or ebpf.CollectionSpec.LoadAndAssign.
type sklookupMaps struct {
	EchoPorts  *ebpf.Map `ebpf:"echo_ports"`
	EchoSocket *ebpf.Map `ebpf:"echo_socket"`
}

func (m *sklookupMaps) Close() error {
	return _SklookupClose(
		m.EchoPorts,
		m.EchoSocket,
	)
}

// sklookupPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadSklookupObjects or ebpf.CollectionSpec.LoadAndAssign.
type sklookupPrograms struct {
	EchoDispatch *ebpf.Program `ebpf:"echo_dispatch"`
}

func (p *sklookupPrograms) Close() error {
	return _SklookupClose(
		p.EchoDispatch,
	)
}

func _SklookupClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed sklookup_bpfeb.o
var _SklookupBytes []byte
