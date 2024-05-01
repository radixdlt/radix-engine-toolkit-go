
package radix_engine_toolkit_uniffi

// #include <radix_engine_toolkit_uniffi.h>
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
	"encoding/binary"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)



type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer {
		capacity: C.int(b.Capacity()),
		len: C.int(b.Len()),
		data: (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func( status *C.RustCallStatus) bool {
		C.ffi_radix_engine_toolkit_uniffi_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}


func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes {
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}
	
	return rustCall(func( status *C.RustCallStatus) RustBuffer {
		return C.ffi_radix_engine_toolkit_uniffi_rustbuffer_from_bytes(foreign, status)
	})
}


type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}


func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}


func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}


func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {
        
        (&FfiConverterCallbackInterfaceSigner{}).register();
        uniffiCheckChecksums()
}


func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_radix_engine_toolkit_uniffi_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: UniFFI contract version mismatch")
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_olympia_account_address_from_public_key(uniffiStatus)
	})
	if checksum != 19647 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_olympia_account_address_from_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_public_key_from_olympia_account_address(uniffiStatus)
	})
	if checksum != 45205 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_public_key_from_olympia_account_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_resource_address_from_olympia_resource_address(uniffiStatus)
	})
	if checksum != 11639 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_resource_address_from_olympia_resource_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_account_address_from_olympia_account_address(uniffiStatus)
	})
	if checksum != 24509 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_account_address_from_olympia_account_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_account_address_from_public_key(uniffiStatus)
	})
	if checksum != 36758 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_account_address_from_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_identity_address_from_public_key(uniffiStatus)
	})
	if checksum != 11003 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_identity_address_from_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_signature_non_fungible_global_id_from_public_key(uniffiStatus)
	})
	if checksum != 61146 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_derive_virtual_signature_non_fungible_global_id_from_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_get_build_information(uniffiStatus)
	})
	if checksum != 61037 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_get_build_information: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_get_hash(uniffiStatus)
	})
	if checksum != 23353 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_get_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_get_known_addresses(uniffiStatus)
	})
	if checksum != 16556 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_get_known_addresses: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_manifest_sbor_decode_to_string_representation(uniffiStatus)
	})
	if checksum != 19578 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_manifest_sbor_decode_to_string_representation: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_metadata_sbor_decode(uniffiStatus)
	})
	if checksum != 54114 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_metadata_sbor_decode: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_metadata_sbor_encode(uniffiStatus)
	})
	if checksum != 11090 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_metadata_sbor_encode: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_as_str(uniffiStatus)
	})
	if checksum != 10663 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_from_str(uniffiStatus)
	})
	if checksum != 27404 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_from_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_sbor_decode(uniffiStatus)
	})
	if checksum != 5482 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_sbor_decode: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_sbor_encode(uniffiStatus)
	})
	if checksum != 44017 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_non_fungible_local_id_sbor_encode: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_public_key_fingerprint_from_vec(uniffiStatus)
	})
	if checksum != 41521 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_public_key_fingerprint_from_vec: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_public_key_fingerprint_to_vec(uniffiStatus)
	})
	if checksum != 4950 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_public_key_fingerprint_to_vec: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_sbor_decode_to_string_representation(uniffiStatus)
	})
	if checksum != 11831 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_sbor_decode_to_string_representation: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_sbor_decode_to_typed_native_event(uniffiStatus)
	})
	if checksum != 43789 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_sbor_decode_to_typed_native_event: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_scrypto_sbor_decode_to_string_representation(uniffiStatus)
	})
	if checksum != 50232 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_scrypto_sbor_decode_to_string_representation: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_scrypto_sbor_encode_string_representation(uniffiStatus)
	})
	if checksum != 24947 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_scrypto_sbor_encode_string_representation: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_func_test_panic(uniffiStatus)
	})
	if checksum != 25407 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_func_test_panic: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_accessrule_and(uniffiStatus)
	})
	if checksum != 5785 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_accessrule_and: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_accessrule_or(uniffiStatus)
	})
	if checksum != 27266 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_accessrule_or: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_address_string(uniffiStatus)
	})
	if checksum != 5709 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_address_string: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_as_str(uniffiStatus)
	})
	if checksum != 38197 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_bytes(uniffiStatus)
	})
	if checksum != 16699 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_entity_type(uniffiStatus)
	})
	if checksum != 40172 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_entity_type: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global(uniffiStatus)
	})
	if checksum != 25808 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_component(uniffiStatus)
	})
	if checksum != 58252 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_component: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_consensus_manager(uniffiStatus)
	})
	if checksum != 48841 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_consensus_manager: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_fungible_resource_manager(uniffiStatus)
	})
	if checksum != 55847 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_fungible_resource_manager: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_non_fungible_resource_manager(uniffiStatus)
	})
	if checksum != 16959 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_non_fungible_resource_manager: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_package(uniffiStatus)
	})
	if checksum != 10761 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_package: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_resource_manager(uniffiStatus)
	})
	if checksum != 34705 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_resource_manager: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_virtual(uniffiStatus)
	})
	if checksum != 44552 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_global_virtual: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal(uniffiStatus)
	})
	if checksum != 34745 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_fungible_vault(uniffiStatus)
	})
	if checksum != 26605 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_fungible_vault: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_kv_store(uniffiStatus)
	})
	if checksum != 4366 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_kv_store: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_non_fungible_vault(uniffiStatus)
	})
	if checksum != 30524 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_non_fungible_vault: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_vault(uniffiStatus)
	})
	if checksum != 10507 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_is_internal_vault: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_address_network_id(uniffiStatus)
	})
	if checksum != 20026 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_address_network_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_abs(uniffiStatus)
	})
	if checksum != 31072 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_abs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_add(uniffiStatus)
	})
	if checksum != 42883 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_add: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_as_str(uniffiStatus)
	})
	if checksum != 18253 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_cbrt(uniffiStatus)
	})
	if checksum != 18756 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_cbrt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_ceiling(uniffiStatus)
	})
	if checksum != 62165 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_ceiling: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_div(uniffiStatus)
	})
	if checksum != 25038 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_div: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_equal(uniffiStatus)
	})
	if checksum != 45597 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_floor(uniffiStatus)
	})
	if checksum != 31716 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_floor: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_greater_than(uniffiStatus)
	})
	if checksum != 16609 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_greater_than: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_greater_than_or_equal(uniffiStatus)
	})
	if checksum != 3170 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_greater_than_or_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_is_negative(uniffiStatus)
	})
	if checksum != 27762 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_is_negative: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_is_positive(uniffiStatus)
	})
	if checksum != 15349 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_is_positive: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_is_zero(uniffiStatus)
	})
	if checksum != 27694 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_is_zero: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_less_than(uniffiStatus)
	})
	if checksum != 30546 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_less_than: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_less_than_or_equal(uniffiStatus)
	})
	if checksum != 2387 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_less_than_or_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_mantissa(uniffiStatus)
	})
	if checksum != 41794 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_mantissa: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_mul(uniffiStatus)
	})
	if checksum != 18912 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_mul: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_not_equal(uniffiStatus)
	})
	if checksum != 61801 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_not_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_nth_root(uniffiStatus)
	})
	if checksum != 6178 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_nth_root: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_powi(uniffiStatus)
	})
	if checksum != 35861 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_powi: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_round(uniffiStatus)
	})
	if checksum != 31873 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_round: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_sqrt(uniffiStatus)
	})
	if checksum != 43295 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_sqrt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_sub(uniffiStatus)
	})
	if checksum != 26365 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_sub: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_to_le_bytes(uniffiStatus)
	})
	if checksum != 17037 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_decimal_to_le_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_hash_as_str(uniffiStatus)
	})
	if checksum != 46597 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_hash_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_hash_bytes(uniffiStatus)
	})
	if checksum != 57303 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_hash_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_instructions_as_str(uniffiStatus)
	})
	if checksum != 2403 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_instructions_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_instructions_instructions_list(uniffiStatus)
	})
	if checksum != 45845 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_instructions_instructions_list: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_instructions_network_id(uniffiStatus)
	})
	if checksum != 55489 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_instructions_network_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_compile(uniffiStatus)
	})
	if checksum != 31325 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_compile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_hash(uniffiStatus)
	})
	if checksum != 993 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_header(uniffiStatus)
	})
	if checksum != 49719 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_header: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_intent_hash(uniffiStatus)
	})
	if checksum != 63530 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_intent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_manifest(uniffiStatus)
	})
	if checksum != 60823 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_manifest: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_message(uniffiStatus)
	})
	if checksum != 49610 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_message: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_statically_validate(uniffiStatus)
	})
	if checksum != 18502 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_intent_statically_validate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_primary_role_badge_withdraw_attempt(uniffiStatus)
	})
	if checksum != 48569 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_primary_role_badge_withdraw_attempt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_primary_role_recovery_proposal(uniffiStatus)
	})
	if checksum != 15034 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_primary_role_recovery_proposal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_recovery_role_badge_withdraw_attempt(uniffiStatus)
	})
	if checksum != 302 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_recovery_role_badge_withdraw_attempt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_recovery_role_recovery_proposal(uniffiStatus)
	})
	if checksum != 29975 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_cancel_recovery_role_recovery_proposal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_create(uniffiStatus)
	})
	if checksum != 58316 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_create: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_create_proof(uniffiStatus)
	})
	if checksum != 64981 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_create_proof: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_create_with_security_structure(uniffiStatus)
	})
	if checksum != 28637 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_create_with_security_structure: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_badge_withdraw_as_primary(uniffiStatus)
	})
	if checksum != 61645 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_badge_withdraw_as_primary: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_badge_withdraw_as_recovery(uniffiStatus)
	})
	if checksum != 57712 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_badge_withdraw_as_recovery: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_recovery_as_primary(uniffiStatus)
	})
	if checksum != 20119 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_recovery_as_primary: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_recovery_as_recovery(uniffiStatus)
	})
	if checksum != 33445 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_initiate_recovery_as_recovery: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_lock_primary_role(uniffiStatus)
	})
	if checksum != 31780 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_lock_primary_role: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_mint_recovery_badges(uniffiStatus)
	})
	if checksum != 4851 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_mint_recovery_badges: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_new_from_public_keys(uniffiStatus)
	})
	if checksum != 6146 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_new_from_public_keys: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_primary_role_badge_withdraw_attempt(uniffiStatus)
	})
	if checksum != 12412 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_primary_role_badge_withdraw_attempt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_primary_role_recovery_proposal(uniffiStatus)
	})
	if checksum != 45088 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_primary_role_recovery_proposal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_recovery_role_badge_withdraw_attempt(uniffiStatus)
	})
	if checksum != 17833 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_recovery_role_badge_withdraw_attempt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_recovery_role_recovery_proposal(uniffiStatus)
	})
	if checksum != 62781 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_quick_confirm_recovery_role_recovery_proposal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_stop_timed_recovery(uniffiStatus)
	})
	if checksum != 34245 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_stop_timed_recovery: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_timed_confirm_recovery(uniffiStatus)
	})
	if checksum != 45733 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_timed_confirm_recovery: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_unlock_primary_role(uniffiStatus)
	})
	if checksum != 30029 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_access_controller_unlock_primary_role: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_add_authorized_depositor(uniffiStatus)
	})
	if checksum != 59221 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_add_authorized_depositor: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_burn(uniffiStatus)
	})
	if checksum != 64728 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_burn: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_burn_non_fungibles(uniffiStatus)
	})
	if checksum != 40710 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_burn_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create(uniffiStatus)
	})
	if checksum != 6013 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create_advanced(uniffiStatus)
	})
	if checksum != 54940 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create_advanced: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create_proof_of_amount(uniffiStatus)
	})
	if checksum != 17393 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create_proof_of_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create_proof_of_non_fungibles(uniffiStatus)
	})
	if checksum != 43091 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_create_proof_of_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_deposit(uniffiStatus)
	})
	if checksum != 3687 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_deposit: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_deposit_batch(uniffiStatus)
	})
	if checksum != 43520 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_deposit_batch: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_deposit_entire_worktop(uniffiStatus)
	})
	if checksum != 59635 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_deposit_entire_worktop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_contingent_fee(uniffiStatus)
	})
	if checksum != 54668 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_contingent_fee: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_fee(uniffiStatus)
	})
	if checksum != 38082 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_fee: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_fee_and_withdraw(uniffiStatus)
	})
	if checksum != 19367 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_fee_and_withdraw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_fee_and_withdraw_non_fungibles(uniffiStatus)
	})
	if checksum != 46012 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_lock_fee_and_withdraw_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_airdrop(uniffiStatus)
	})
	if checksum != 40671 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_airdrop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_claim(uniffiStatus)
	})
	if checksum != 20662 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_claim: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_claim_non_fungibles(uniffiStatus)
	})
	if checksum != 38683 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_claim_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_get_amount(uniffiStatus)
	})
	if checksum != 29397 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_get_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_get_non_fungible_local_ids(uniffiStatus)
	})
	if checksum != 23927 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_get_non_fungible_local_ids: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_instantiate(uniffiStatus)
	})
	if checksum != 7727 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_instantiate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_instantiate_simple(uniffiStatus)
	})
	if checksum != 65307 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_instantiate_simple: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_recover(uniffiStatus)
	})
	if checksum != 2999 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_recover: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_recover_non_fungibles(uniffiStatus)
	})
	if checksum != 65115 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_recover_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_store(uniffiStatus)
	})
	if checksum != 40050 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_locker_store: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_remove_authorized_depositor(uniffiStatus)
	})
	if checksum != 17654 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_remove_authorized_depositor: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_remove_resource_preference(uniffiStatus)
	})
	if checksum != 57432 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_remove_resource_preference: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_securify(uniffiStatus)
	})
	if checksum != 20811 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_securify: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_set_default_deposit_rule(uniffiStatus)
	})
	if checksum != 28798 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_set_default_deposit_rule: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_set_resource_preference(uniffiStatus)
	})
	if checksum != 24940 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_set_resource_preference: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_batch_or_abort(uniffiStatus)
	})
	if checksum != 18649 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_batch_or_abort: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_batch_or_refund(uniffiStatus)
	})
	if checksum != 34909 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_batch_or_refund: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_entire_worktop_or_abort(uniffiStatus)
	})
	if checksum != 42658 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_entire_worktop_or_abort: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_entire_worktop_or_refund(uniffiStatus)
	})
	if checksum != 9020 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_entire_worktop_or_refund: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_or_abort(uniffiStatus)
	})
	if checksum != 21998 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_or_abort: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_or_refund(uniffiStatus)
	})
	if checksum != 39086 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_try_deposit_or_refund: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_withdraw(uniffiStatus)
	})
	if checksum != 29156 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_withdraw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_withdraw_non_fungibles(uniffiStatus)
	})
	if checksum != 56678 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_account_withdraw_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_allocate_global_address(uniffiStatus)
	})
	if checksum != 18604 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_allocate_global_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_assert_worktop_contains(uniffiStatus)
	})
	if checksum != 37738 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_assert_worktop_contains: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_assert_worktop_contains_any(uniffiStatus)
	})
	if checksum != 20665 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_assert_worktop_contains_any: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_assert_worktop_contains_non_fungibles(uniffiStatus)
	})
	if checksum != 58282 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_assert_worktop_contains_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_build(uniffiStatus)
	})
	if checksum != 36705 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_build: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_burn_resource(uniffiStatus)
	})
	if checksum != 52445 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_burn_resource: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_access_rules_method(uniffiStatus)
	})
	if checksum != 19399 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_access_rules_method: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_direct_vault_method(uniffiStatus)
	})
	if checksum != 53674 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_direct_vault_method: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_function(uniffiStatus)
	})
	if checksum != 38619 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_function: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_metadata_method(uniffiStatus)
	})
	if checksum != 42239 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_metadata_method: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_method(uniffiStatus)
	})
	if checksum != 39370 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_method: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_royalty_method(uniffiStatus)
	})
	if checksum != 25488 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_call_royalty_method: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_clone_proof(uniffiStatus)
	})
	if checksum != 52407 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_clone_proof: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_fungible_resource_manager(uniffiStatus)
	})
	if checksum != 45955 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_fungible_resource_manager: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_auth_zone_of_all(uniffiStatus)
	})
	if checksum != 51538 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_auth_zone_of_all: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_auth_zone_of_amount(uniffiStatus)
	})
	if checksum != 51265 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_auth_zone_of_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_auth_zone_of_non_fungibles(uniffiStatus)
	})
	if checksum != 49166 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_auth_zone_of_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_bucket_of_all(uniffiStatus)
	})
	if checksum != 46129 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_bucket_of_all: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_bucket_of_amount(uniffiStatus)
	})
	if checksum != 20827 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_bucket_of_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_bucket_of_non_fungibles(uniffiStatus)
	})
	if checksum != 25333 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_create_proof_from_bucket_of_non_fungibles: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_all_proofs(uniffiStatus)
	})
	if checksum != 12341 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_all_proofs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_auth_zone_proofs(uniffiStatus)
	})
	if checksum != 63484 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_auth_zone_proofs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_auth_zone_signature_proofs(uniffiStatus)
	})
	if checksum != 2952 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_auth_zone_signature_proofs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_proof(uniffiStatus)
	})
	if checksum != 29894 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_drop_proof: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_faucet_free_xrd(uniffiStatus)
	})
	if checksum != 59721 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_faucet_free_xrd: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_faucet_lock_fee(uniffiStatus)
	})
	if checksum != 5856 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_faucet_lock_fee: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_identity_create(uniffiStatus)
	})
	if checksum != 22657 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_identity_create: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_identity_create_advanced(uniffiStatus)
	})
	if checksum != 53046 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_identity_create_advanced: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_identity_securify(uniffiStatus)
	})
	if checksum != 24322 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_identity_securify: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_get(uniffiStatus)
	})
	if checksum != 37782 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_get: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_lock(uniffiStatus)
	})
	if checksum != 53375 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_lock: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_remove(uniffiStatus)
	})
	if checksum != 30456 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_remove: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_set(uniffiStatus)
	})
	if checksum != 11186 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_metadata_set: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_mint_fungible(uniffiStatus)
	})
	if checksum != 41635 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_mint_fungible: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_contribute(uniffiStatus)
	})
	if checksum != 54648 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_contribute: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_get_redemption_value(uniffiStatus)
	})
	if checksum != 1278 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_get_redemption_value: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_get_vault_amount(uniffiStatus)
	})
	if checksum != 53964 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_get_vault_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_instantiate(uniffiStatus)
	})
	if checksum != 17825 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_instantiate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_protected_deposit(uniffiStatus)
	})
	if checksum != 10939 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_protected_deposit: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_protected_withdraw(uniffiStatus)
	})
	if checksum != 3505 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_protected_withdraw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_redeem(uniffiStatus)
	})
	if checksum != 16912 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_multi_resource_pool_redeem: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_contribute(uniffiStatus)
	})
	if checksum != 25120 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_contribute: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_get_redemption_value(uniffiStatus)
	})
	if checksum != 27814 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_get_redemption_value: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_get_vault_amount(uniffiStatus)
	})
	if checksum != 37942 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_get_vault_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_instantiate(uniffiStatus)
	})
	if checksum != 5474 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_instantiate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_protected_deposit(uniffiStatus)
	})
	if checksum != 1325 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_protected_deposit: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_protected_withdraw(uniffiStatus)
	})
	if checksum != 47007 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_protected_withdraw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_redeem(uniffiStatus)
	})
	if checksum != 16139 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_one_resource_pool_redeem: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_package_claim_royalty(uniffiStatus)
	})
	if checksum != 54897 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_package_claim_royalty: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_package_publish(uniffiStatus)
	})
	if checksum != 43039 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_package_publish: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_package_publish_advanced(uniffiStatus)
	})
	if checksum != 7234 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_package_publish_advanced: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_pop_from_auth_zone(uniffiStatus)
	})
	if checksum != 54385 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_pop_from_auth_zone: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_push_to_auth_zone(uniffiStatus)
	})
	if checksum != 59668 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_push_to_auth_zone: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_return_to_worktop(uniffiStatus)
	})
	if checksum != 48542 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_return_to_worktop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_get(uniffiStatus)
	})
	if checksum != 57962 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_get: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_lock_owner(uniffiStatus)
	})
	if checksum != 26186 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_lock_owner: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_set(uniffiStatus)
	})
	if checksum != 27207 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_set: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_set_owner(uniffiStatus)
	})
	if checksum != 64161 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_role_assignment_set_owner: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_royalty_claim(uniffiStatus)
	})
	if checksum != 23601 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_royalty_claim: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_royalty_lock(uniffiStatus)
	})
	if checksum != 50599 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_royalty_lock: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_royalty_set(uniffiStatus)
	})
	if checksum != 26584 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_royalty_set: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_take_all_from_worktop(uniffiStatus)
	})
	if checksum != 61948 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_take_all_from_worktop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_take_from_worktop(uniffiStatus)
	})
	if checksum != 7334 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_take_from_worktop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_take_non_fungibles_from_worktop(uniffiStatus)
	})
	if checksum != 49676 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_take_non_fungibles_from_worktop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_contribute(uniffiStatus)
	})
	if checksum != 3256 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_contribute: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_get_redemption_value(uniffiStatus)
	})
	if checksum != 41038 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_get_redemption_value: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_get_vault_amount(uniffiStatus)
	})
	if checksum != 44545 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_get_vault_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_instantiate(uniffiStatus)
	})
	if checksum != 22784 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_instantiate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_protected_deposit(uniffiStatus)
	})
	if checksum != 8937 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_protected_deposit: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_protected_withdraw(uniffiStatus)
	})
	if checksum != 35351 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_protected_withdraw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_redeem(uniffiStatus)
	})
	if checksum != 4503 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_two_resource_pool_redeem: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_accepts_delegated_stake(uniffiStatus)
	})
	if checksum != 63411 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_accepts_delegated_stake: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_claim_xrd(uniffiStatus)
	})
	if checksum != 13361 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_claim_xrd: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_finish_unlock_owner_stake_units(uniffiStatus)
	})
	if checksum != 24114 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_finish_unlock_owner_stake_units: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_get_protocol_update_readiness(uniffiStatus)
	})
	if checksum != 56572 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_get_protocol_update_readiness: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_get_redemption_value(uniffiStatus)
	})
	if checksum != 30890 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_get_redemption_value: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_lock_owner_stake_units(uniffiStatus)
	})
	if checksum != 26840 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_lock_owner_stake_units: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_register(uniffiStatus)
	})
	if checksum != 38592 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_register: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_signal_protocol_update_readiness(uniffiStatus)
	})
	if checksum != 41037 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_signal_protocol_update_readiness: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_stake(uniffiStatus)
	})
	if checksum != 46849 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_stake: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_stake_as_owner(uniffiStatus)
	})
	if checksum != 43974 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_stake_as_owner: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_start_unlock_owner_stake_units(uniffiStatus)
	})
	if checksum != 53351 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_start_unlock_owner_stake_units: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_total_stake_unit_supply(uniffiStatus)
	})
	if checksum != 14885 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_total_stake_unit_supply: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_total_stake_xrd_amount(uniffiStatus)
	})
	if checksum != 44141 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_total_stake_xrd_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_unregister(uniffiStatus)
	})
	if checksum != 55641 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_unregister: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_unstake(uniffiStatus)
	})
	if checksum != 53557 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_unstake: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_update_accept_delegated_stake(uniffiStatus)
	})
	if checksum != 38363 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_update_accept_delegated_stake: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_update_fee(uniffiStatus)
	})
	if checksum != 10602 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_update_fee: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_update_key(uniffiStatus)
	})
	if checksum != 41122 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_manifestbuilder_validator_update_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_decryptors(uniffiStatus)
	})
	if checksum != 45350 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_decryptors: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_encrypted_message_length(uniffiStatus)
	})
	if checksum != 10753 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_encrypted_message_length: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_mime_type_length(uniffiStatus)
	})
	if checksum != 15824 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_mime_type_length: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_plaintext_message_length(uniffiStatus)
	})
	if checksum != 53437 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_messagevalidationconfig_max_plaintext_message_length: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_nonfungibleglobalid_as_str(uniffiStatus)
	})
	if checksum != 12617 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_nonfungibleglobalid_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_nonfungibleglobalid_local_id(uniffiStatus)
	})
	if checksum != 42729 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_nonfungibleglobalid_local_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_nonfungibleglobalid_resource_address(uniffiStatus)
	})
	if checksum != 26038 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_nonfungibleglobalid_resource_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_compile(uniffiStatus)
	})
	if checksum != 65183 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_compile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_hash(uniffiStatus)
	})
	if checksum != 64270 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_intent_hash(uniffiStatus)
	})
	if checksum != 51688 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_intent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_notarized_transaction_hash(uniffiStatus)
	})
	if checksum != 17757 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_notarized_transaction_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_notary_signature(uniffiStatus)
	})
	if checksum != 46873 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_notary_signature: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_signed_intent(uniffiStatus)
	})
	if checksum != 11409 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_signed_intent: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_signed_intent_hash(uniffiStatus)
	})
	if checksum != 60604 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_signed_intent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_statically_validate(uniffiStatus)
	})
	if checksum != 11188 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_notarizedtransaction_statically_validate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_olympiaaddress_as_str(uniffiStatus)
	})
	if checksum != 211 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_olympiaaddress_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_olympiaaddress_public_key(uniffiStatus)
	})
	if checksum != 33649 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_olympiaaddress_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_abs(uniffiStatus)
	})
	if checksum != 753 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_abs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_add(uniffiStatus)
	})
	if checksum != 50067 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_add: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_as_str(uniffiStatus)
	})
	if checksum != 50135 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_cbrt(uniffiStatus)
	})
	if checksum != 31353 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_cbrt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_ceiling(uniffiStatus)
	})
	if checksum != 6632 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_ceiling: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_div(uniffiStatus)
	})
	if checksum != 47336 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_div: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_equal(uniffiStatus)
	})
	if checksum != 35658 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_floor(uniffiStatus)
	})
	if checksum != 6297 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_floor: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_greater_than(uniffiStatus)
	})
	if checksum != 21292 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_greater_than: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_greater_than_or_equal(uniffiStatus)
	})
	if checksum != 34931 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_greater_than_or_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_is_negative(uniffiStatus)
	})
	if checksum != 11588 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_is_negative: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_is_positive(uniffiStatus)
	})
	if checksum != 30868 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_is_positive: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_is_zero(uniffiStatus)
	})
	if checksum != 41566 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_is_zero: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_less_than(uniffiStatus)
	})
	if checksum != 50862 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_less_than: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_less_than_or_equal(uniffiStatus)
	})
	if checksum != 33893 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_less_than_or_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_mantissa(uniffiStatus)
	})
	if checksum != 2374 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_mantissa: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_mul(uniffiStatus)
	})
	if checksum != 35568 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_mul: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_not_equal(uniffiStatus)
	})
	if checksum != 17368 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_not_equal: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_nth_root(uniffiStatus)
	})
	if checksum != 60037 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_nth_root: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_powi(uniffiStatus)
	})
	if checksum != 57119 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_powi: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_round(uniffiStatus)
	})
	if checksum != 22122 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_round: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_sqrt(uniffiStatus)
	})
	if checksum != 18565 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_sqrt: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_sub(uniffiStatus)
	})
	if checksum != 2969 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_sub: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_to_le_bytes(uniffiStatus)
	})
	if checksum != 6841 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_precisedecimal_to_le_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_curve(uniffiStatus)
	})
	if checksum != 56035 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_curve: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_public_key(uniffiStatus)
	})
	if checksum != 49403 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_public_key_bytes(uniffiStatus)
	})
	if checksum != 8464 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_public_key_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_raw(uniffiStatus)
	})
	if checksum != 43216 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_raw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_raw_hex(uniffiStatus)
	})
	if checksum != 64460 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_raw_hex: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_sign(uniffiStatus)
	})
	if checksum != 21427 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_sign: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_sign_to_signature(uniffiStatus)
	})
	if checksum != 4246 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_sign_to_signature: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_sign_to_signature_with_public_key(uniffiStatus)
	})
	if checksum != 41168 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_privatekey_sign_to_signature_with_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_compile(uniffiStatus)
	})
	if checksum != 26394 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_compile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_hash(uniffiStatus)
	})
	if checksum != 60260 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_intent(uniffiStatus)
	})
	if checksum != 19540 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_intent: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_intent_hash(uniffiStatus)
	})
	if checksum != 9462 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_intent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_intent_signatures(uniffiStatus)
	})
	if checksum != 46037 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_intent_signatures: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_signed_intent_hash(uniffiStatus)
	})
	if checksum != 20757 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_signed_intent_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_statically_validate(uniffiStatus)
	})
	if checksum != 27682 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signedintent_statically_validate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilder_header(uniffiStatus)
	})
	if checksum != 40383 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilder_header: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderheaderstep_manifest(uniffiStatus)
	})
	if checksum != 8446 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderheaderstep_manifest: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_notarize_with_private_key(uniffiStatus)
	})
	if checksum != 57025 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_notarize_with_private_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_notarize_with_signer(uniffiStatus)
	})
	if checksum != 32547 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_notarize_with_signer: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_sign_with_private_key(uniffiStatus)
	})
	if checksum != 29671 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_sign_with_private_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_sign_with_signer(uniffiStatus)
	})
	if checksum != 17372 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuilderintentsignaturesstep_sign_with_signer: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuildermessagestep_message(uniffiStatus)
	})
	if checksum != 55782 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuildermessagestep_message: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuildermessagestep_sign_with_private_key(uniffiStatus)
	})
	if checksum != 60073 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuildermessagestep_sign_with_private_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuildermessagestep_sign_with_signer(uniffiStatus)
	})
	if checksum != 21713 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionbuildermessagestep_sign_with_signer: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_as_hash(uniffiStatus)
	})
	if checksum != 1343 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_as_hash: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_as_str(uniffiStatus)
	})
	if checksum != 9829 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_as_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_bytes(uniffiStatus)
	})
	if checksum != 40875 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_network_id(uniffiStatus)
	})
	if checksum != 4187 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionhash_network_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_blobs(uniffiStatus)
	})
	if checksum != 55127 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_blobs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_compile(uniffiStatus)
	})
	if checksum != 11452 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_compile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_execution_summary(uniffiStatus)
	})
	if checksum != 43934 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_execution_summary: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_extract_addresses(uniffiStatus)
	})
	if checksum != 5474 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_extract_addresses: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_instructions(uniffiStatus)
	})
	if checksum != 3783 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_instructions: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_modify(uniffiStatus)
	})
	if checksum != 4850 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_modify: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_statically_validate(uniffiStatus)
	})
	if checksum != 42656 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_statically_validate: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_summary(uniffiStatus)
	})
	if checksum != 53923 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_transactionmanifest_summary: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_max_epoch_range(uniffiStatus)
	})
	if checksum != 31430 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_max_epoch_range: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_max_notarized_payload_size(uniffiStatus)
	})
	if checksum != 39564 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_max_notarized_payload_size: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_max_tip_percentage(uniffiStatus)
	})
	if checksum != 28981 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_max_tip_percentage: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_message_validation(uniffiStatus)
	})
	if checksum != 52946 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_message_validation: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_min_tip_percentage(uniffiStatus)
	})
	if checksum != 2069 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_min_tip_percentage: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_network_id(uniffiStatus)
	})
	if checksum != 63098 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_validationconfig_network_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_allow_all(uniffiStatus)
	})
	if checksum != 26074 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_allow_all: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_deny_all(uniffiStatus)
	})
	if checksum != 40312 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_deny_all: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require(uniffiStatus)
	})
	if checksum != 10110 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_all_of(uniffiStatus)
	})
	if checksum != 11748 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_all_of: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_amount(uniffiStatus)
	})
	if checksum != 34714 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_amount: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_any_of(uniffiStatus)
	})
	if checksum != 30352 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_any_of: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_count_of(uniffiStatus)
	})
	if checksum != 59472 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_count_of: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_virtual_signature(uniffiStatus)
	})
	if checksum != 41270 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_accessrule_require_virtual_signature: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_from_raw(uniffiStatus)
	})
	if checksum != 43797 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_from_raw: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_new(uniffiStatus)
	})
	if checksum != 37549 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_resource_address_from_olympia_resource_address(uniffiStatus)
	})
	if checksum != 64771 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_resource_address_from_olympia_resource_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_virtual_account_address_from_olympia_address(uniffiStatus)
	})
	if checksum != 31070 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_virtual_account_address_from_olympia_address: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_virtual_account_address_from_public_key(uniffiStatus)
	})
	if checksum != 738 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_virtual_account_address_from_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_virtual_identity_address_from_public_key(uniffiStatus)
	})
	if checksum != 32432 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_address_virtual_identity_address_from_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_from_le_bytes(uniffiStatus)
	})
	if checksum != 14760 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_from_le_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_max(uniffiStatus)
	})
	if checksum != 38313 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_max: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_min(uniffiStatus)
	})
	if checksum != 18079 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_min: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_new(uniffiStatus)
	})
	if checksum != 15617 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_one(uniffiStatus)
	})
	if checksum != 42470 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_one: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_zero(uniffiStatus)
	})
	if checksum != 39451 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_decimal_zero: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_from_hex_string(uniffiStatus)
	})
	if checksum != 64410 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_from_hex_string: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_from_unhashed_bytes(uniffiStatus)
	})
	if checksum != 17030 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_from_unhashed_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_new(uniffiStatus)
	})
	if checksum != 17594 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_sbor_decode(uniffiStatus)
	})
	if checksum != 26443 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_hash_sbor_decode: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_instructions_from_instructions(uniffiStatus)
	})
	if checksum != 51039 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_instructions_from_instructions: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_instructions_from_string(uniffiStatus)
	})
	if checksum != 47420 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_instructions_from_string: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_intent_decompile(uniffiStatus)
	})
	if checksum != 565 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_intent_decompile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_intent_new(uniffiStatus)
	})
	if checksum != 4284 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_intent_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_manifestbuilder_new(uniffiStatus)
	})
	if checksum != 30710 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_manifestbuilder_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_messagevalidationconfig_default(uniffiStatus)
	})
	if checksum != 54905 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_messagevalidationconfig_default: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_messagevalidationconfig_new(uniffiStatus)
	})
	if checksum != 60275 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_messagevalidationconfig_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_nonfungibleglobalid_from_parts(uniffiStatus)
	})
	if checksum != 36478 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_nonfungibleglobalid_from_parts: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_nonfungibleglobalid_new(uniffiStatus)
	})
	if checksum != 58056 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_nonfungibleglobalid_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_nonfungibleglobalid_virtual_signature_badge(uniffiStatus)
	})
	if checksum != 22546 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_nonfungibleglobalid_virtual_signature_badge: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_notarizedtransaction_decompile(uniffiStatus)
	})
	if checksum != 58667 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_notarizedtransaction_decompile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_notarizedtransaction_new(uniffiStatus)
	})
	if checksum != 56154 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_notarizedtransaction_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_olympiaaddress_new(uniffiStatus)
	})
	if checksum != 12724 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_olympiaaddress_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_from_le_bytes(uniffiStatus)
	})
	if checksum != 24547 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_from_le_bytes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_max(uniffiStatus)
	})
	if checksum != 49495 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_max: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_min(uniffiStatus)
	})
	if checksum != 4453 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_min: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_new(uniffiStatus)
	})
	if checksum != 34846 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_one(uniffiStatus)
	})
	if checksum != 9121 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_one: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_zero(uniffiStatus)
	})
	if checksum != 5648 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_precisedecimal_zero: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_privatekey_new(uniffiStatus)
	})
	if checksum != 47612 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_privatekey_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_privatekey_new_ed25519(uniffiStatus)
	})
	if checksum != 4005 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_privatekey_new_ed25519: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_privatekey_new_secp256k1(uniffiStatus)
	})
	if checksum != 20991 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_privatekey_new_secp256k1: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_signedintent_decompile(uniffiStatus)
	})
	if checksum != 12765 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_signedintent_decompile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_signedintent_new(uniffiStatus)
	})
	if checksum != 36392 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_signedintent_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionbuilder_new(uniffiStatus)
	})
	if checksum != 46196 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionbuilder_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionbuilderintentsignaturesstep_new(uniffiStatus)
	})
	if checksum != 17229 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionbuilderintentsignaturesstep_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionhash_from_str(uniffiStatus)
	})
	if checksum != 37610 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionhash_from_str: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionmanifest_decompile(uniffiStatus)
	})
	if checksum != 51209 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionmanifest_decompile: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionmanifest_new(uniffiStatus)
	})
	if checksum != 62865 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_transactionmanifest_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_validationconfig_default(uniffiStatus)
	})
	if checksum != 1435 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_validationconfig_default: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_constructor_validationconfig_new(uniffiStatus)
	})
	if checksum != 36594 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_constructor_validationconfig_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_sign(uniffiStatus)
	})
	if checksum != 46892 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_sign: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_sign_to_signature(uniffiStatus)
	})
	if checksum != 15804 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_sign_to_signature: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_sign_to_signature_with_public_key(uniffiStatus)
	})
	if checksum != 9393 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_sign_to_signature_with_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_public_key(uniffiStatus)
	})
	if checksum != 61195 {
		// If this happens try cleaning and rebuilding your project
		panic("radix_engine_toolkit_uniffi: uniffi_radix_engine_toolkit_uniffi_checksum_method_signer_public_key: UniFFI API checksum mismatch")
	}
	}
}




type FfiConverterUint8 struct{}

var FfiConverterUint8INSTANCE = FfiConverterUint8{}

func (FfiConverterUint8) Lower(value uint8) C.uint8_t {
	return C.uint8_t(value)
}

func (FfiConverterUint8) Write(writer io.Writer, value uint8) {
	writeUint8(writer, value)
}

func (FfiConverterUint8) Lift(value C.uint8_t) uint8 {
	return uint8(value)
}

func (FfiConverterUint8) Read(reader io.Reader) uint8 {
	return readUint8(reader)
}

type FfiDestroyerUint8 struct {}

func (FfiDestroyerUint8) Destroy(_ uint8) {}


type FfiConverterInt8 struct{}

var FfiConverterInt8INSTANCE = FfiConverterInt8{}

func (FfiConverterInt8) Lower(value int8) C.int8_t {
	return C.int8_t(value)
}

func (FfiConverterInt8) Write(writer io.Writer, value int8) {
	writeInt8(writer, value)
}

func (FfiConverterInt8) Lift(value C.int8_t) int8 {
	return int8(value)
}

func (FfiConverterInt8) Read(reader io.Reader) int8 {
	return readInt8(reader)
}

type FfiDestroyerInt8 struct {}

func (FfiDestroyerInt8) Destroy(_ int8) {}


type FfiConverterUint16 struct{}

var FfiConverterUint16INSTANCE = FfiConverterUint16{}

func (FfiConverterUint16) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func (FfiConverterUint16) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func (FfiConverterUint16) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func (FfiConverterUint16) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type FfiDestroyerUint16 struct {}

func (FfiDestroyerUint16) Destroy(_ uint16) {}


type FfiConverterInt16 struct{}

var FfiConverterInt16INSTANCE = FfiConverterInt16{}

func (FfiConverterInt16) Lower(value int16) C.int16_t {
	return C.int16_t(value)
}

func (FfiConverterInt16) Write(writer io.Writer, value int16) {
	writeInt16(writer, value)
}

func (FfiConverterInt16) Lift(value C.int16_t) int16 {
	return int16(value)
}

func (FfiConverterInt16) Read(reader io.Reader) int16 {
	return readInt16(reader)
}

type FfiDestroyerInt16 struct {}

func (FfiDestroyerInt16) Destroy(_ int16) {}


type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct {}

func (FfiDestroyerUint32) Destroy(_ uint32) {}


type FfiConverterInt32 struct{}

var FfiConverterInt32INSTANCE = FfiConverterInt32{}

func (FfiConverterInt32) Lower(value int32) C.int32_t {
	return C.int32_t(value)
}

func (FfiConverterInt32) Write(writer io.Writer, value int32) {
	writeInt32(writer, value)
}

func (FfiConverterInt32) Lift(value C.int32_t) int32 {
	return int32(value)
}

func (FfiConverterInt32) Read(reader io.Reader) int32 {
	return readInt32(reader)
}

type FfiDestroyerInt32 struct {}

func (FfiDestroyerInt32) Destroy(_ int32) {}


type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct {}

func (FfiDestroyerUint64) Destroy(_ uint64) {}


type FfiConverterInt64 struct{}

var FfiConverterInt64INSTANCE = FfiConverterInt64{}

func (FfiConverterInt64) Lower(value int64) C.int64_t {
	return C.int64_t(value)
}

func (FfiConverterInt64) Write(writer io.Writer, value int64) {
	writeInt64(writer, value)
}

func (FfiConverterInt64) Lift(value C.int64_t) int64 {
	return int64(value)
}

func (FfiConverterInt64) Read(reader io.Reader) int64 {
	return readInt64(reader)
}

type FfiDestroyerInt64 struct {}

func (FfiDestroyerInt64) Destroy(_ int64) {}


type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct {}

func (FfiDestroyerBool) Destroy(_ bool) {}


type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct {}

func (FfiDestroyerString) Destroy(_ string) {}


type FfiConverterBytes struct{}

var FfiConverterBytesINSTANCE = FfiConverterBytes{}

func (c FfiConverterBytes) Lower(value []byte) RustBuffer {
	return LowerIntoRustBuffer[[]byte](c, value)
}

func (c FfiConverterBytes) Write(writer io.Writer, value []byte) {
	if len(value) > math.MaxInt32 {
		panic("[]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := writer.Write(value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing []byte, expected %d, written %d", len(value), write_length))
	}
}

func (c FfiConverterBytes) Lift(rb RustBufferI) []byte {
	return LiftFromRustBuffer[[]byte](c, rb)
}

func (c FfiConverterBytes) Read(reader io.Reader) []byte {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading []byte, expected %d, read %d", length, read_length))
	}
	return buffer
}

type FfiDestroyerBytes struct {}

func (FfiDestroyerBytes) Destroy(_ []byte) {}



// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer unsafe.Pointer
	callCounter atomic.Int64
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed atomic.Bool
}

func newFfiObject(pointer unsafe.Pointer, freeFunction func(unsafe.Pointer, *C.RustCallStatus)) FfiObject {
	return FfiObject {
		pointer: pointer,
		freeFunction: freeFunction,
	}
}

func (ffiObject *FfiObject)incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter + 1) {
			break
		}
	}

	return ffiObject.pointer
}

func (ffiObject *FfiObject)decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject)destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject)freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}
type AccessRule struct {
	ffiObject FfiObject
}


func AccessRuleAllowAll() *AccessRule {
	return FfiConverterAccessRuleINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_allow_all( _uniffiStatus)
	}))
}

func AccessRuleDenyAll() *AccessRule {
	return FfiConverterAccessRuleINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_deny_all( _uniffiStatus)
	}))
}

func AccessRuleRequire(resourceOrNonFungible ResourceOrNonFungible) (*AccessRule, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_require(FfiConverterTypeResourceOrNonFungibleINSTANCE.Lower(resourceOrNonFungible), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *AccessRule
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAccessRuleINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AccessRuleRequireAllOf(resources []ResourceOrNonFungible) (*AccessRule, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_require_all_of(FfiConverterSequenceTypeResourceOrNonFungibleINSTANCE.Lower(resources), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *AccessRule
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAccessRuleINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AccessRuleRequireAmount(amount *Decimal, resource *Address) (*AccessRule, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_require_amount(FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterAddressINSTANCE.Lower(resource), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *AccessRule
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAccessRuleINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AccessRuleRequireAnyOf(resources []ResourceOrNonFungible) (*AccessRule, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_require_any_of(FfiConverterSequenceTypeResourceOrNonFungibleINSTANCE.Lower(resources), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *AccessRule
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAccessRuleINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AccessRuleRequireCountOf(count uint8, resources []ResourceOrNonFungible) (*AccessRule, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_require_count_of(FfiConverterUint8INSTANCE.Lower(count), FfiConverterSequenceTypeResourceOrNonFungibleINSTANCE.Lower(resources), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *AccessRule
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAccessRuleINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AccessRuleRequireVirtualSignature(publicKey PublicKey) (*AccessRule, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_accessrule_require_virtual_signature(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *AccessRule
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAccessRuleINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *AccessRule)And(other *AccessRule) *AccessRule {
	_pointer := _self.ffiObject.incrementPointer("*AccessRule")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterAccessRuleINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_accessrule_and(
		_pointer,FfiConverterAccessRuleINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *AccessRule)Or(other *AccessRule) *AccessRule {
	_pointer := _self.ffiObject.incrementPointer("*AccessRule")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterAccessRuleINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_accessrule_or(
		_pointer,FfiConverterAccessRuleINSTANCE.Lower(other), _uniffiStatus)
	}))
}



func (object *AccessRule)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterAccessRule struct {}

var FfiConverterAccessRuleINSTANCE = FfiConverterAccessRule{}

func (c FfiConverterAccessRule) Lift(pointer unsafe.Pointer) *AccessRule {
	result := &AccessRule {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_accessrule(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*AccessRule).Destroy)
	return result
}

func (c FfiConverterAccessRule) Read(reader io.Reader) *AccessRule {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterAccessRule) Lower(value *AccessRule) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*AccessRule")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterAccessRule) Write(writer io.Writer, value *AccessRule) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerAccessRule struct {}

func (_ FfiDestroyerAccessRule) Destroy(value *AccessRule) {
	value.Destroy()
}


type Address struct {
	ffiObject FfiObject
}
func NewAddress(address string) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_address_new(FfiConverterStringINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func AddressFromRaw(nodeIdBytes []byte, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_address_from_raw(FfiConverterBytesINSTANCE.Lower(nodeIdBytes), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AddressResourceAddressFromOlympiaResourceAddress(olympiaResourceAddress *OlympiaAddress, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_address_resource_address_from_olympia_resource_address(FfiConverterOlympiaAddressINSTANCE.Lower(olympiaResourceAddress), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AddressVirtualAccountAddressFromOlympiaAddress(olympiaAccountAddress *OlympiaAddress, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_address_virtual_account_address_from_olympia_address(FfiConverterOlympiaAddressINSTANCE.Lower(olympiaAccountAddress), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AddressVirtualAccountAddressFromPublicKey(publicKey PublicKey, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_address_virtual_account_address_from_public_key(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func AddressVirtualIdentityAddressFromPublicKey(publicKey PublicKey, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_address_virtual_identity_address_from_public_key(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *Address)AddressString() string {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_address_string(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)Bytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_bytes(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)EntityType() EntityType {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeEntityTypeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_entity_type(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobal() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalComponent() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_component(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalConsensusManager() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_consensus_manager(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalFungibleResourceManager() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_fungible_resource_manager(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalNonFungibleResourceManager() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_non_fungible_resource_manager(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalPackage() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_package(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalResourceManager() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_resource_manager(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsGlobalVirtual() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_global_virtual(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsInternal() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_internal(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsInternalFungibleVault() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_internal_fungible_vault(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsInternalKvStore() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_internal_kv_store(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsInternalNonFungibleVault() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_internal_non_fungible_vault(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)IsInternalVault() bool {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_is_internal_vault(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Address)NetworkId() uint8 {
	_pointer := _self.ffiObject.incrementPointer("*Address")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint8INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_address_network_id(
		_pointer, _uniffiStatus)
	}))
}



func (object *Address)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterAddress struct {}

var FfiConverterAddressINSTANCE = FfiConverterAddress{}

func (c FfiConverterAddress) Lift(pointer unsafe.Pointer) *Address {
	result := &Address {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_address(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*Address).Destroy)
	return result
}

func (c FfiConverterAddress) Read(reader io.Reader) *Address {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterAddress) Lower(value *Address) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Address")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterAddress) Write(writer io.Writer, value *Address) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerAddress struct {}

func (_ FfiDestroyerAddress) Destroy(value *Address) {
	value.Destroy()
}


type Decimal struct {
	ffiObject FfiObject
}
func NewDecimal(value string) (*Decimal, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_decimal_new(FfiConverterStringINSTANCE.Lower(value), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func DecimalFromLeBytes(value []byte) *Decimal {
	return FfiConverterDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_decimal_from_le_bytes(FfiConverterBytesINSTANCE.Lower(value), _uniffiStatus)
	}))
}

func DecimalMax() *Decimal {
	return FfiConverterDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_decimal_max( _uniffiStatus)
	}))
}

func DecimalMin() *Decimal {
	return FfiConverterDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_decimal_min( _uniffiStatus)
	}))
}

func DecimalOne() *Decimal {
	return FfiConverterDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_decimal_one( _uniffiStatus)
	}))
}

func DecimalZero() *Decimal {
	return FfiConverterDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_decimal_zero( _uniffiStatus)
	}))
}



func (_self *Decimal)Abs() (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_abs(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)Add(other *Decimal) (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_add(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Decimal)Cbrt() (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_cbrt(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)Ceiling() (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_ceiling(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)Div(other *Decimal) (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_div(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)Equal(other *Decimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_equal(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *Decimal)Floor() (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_floor(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)GreaterThan(other *Decimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_greater_than(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *Decimal)GreaterThanOrEqual(other *Decimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_greater_than_or_equal(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *Decimal)IsNegative() bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_is_negative(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Decimal)IsPositive() bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_is_positive(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Decimal)IsZero() bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_is_zero(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Decimal)LessThan(other *Decimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_less_than(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *Decimal)LessThanOrEqual(other *Decimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_less_than_or_equal(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *Decimal)Mantissa() string {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_mantissa(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Decimal)Mul(other *Decimal) (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_mul(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)NotEqual(other *Decimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_not_equal(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *Decimal)NthRoot(n uint32) **Decimal {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_nth_root(
		_pointer,FfiConverterUint32INSTANCE.Lower(n), _uniffiStatus)
	}))
}


func (_self *Decimal)Powi(exp int64) (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_powi(
		_pointer,FfiConverterInt64INSTANCE.Lower(exp), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)Round(decimalPlaces int32, roundingMode RoundingMode) (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_round(
		_pointer,FfiConverterInt32INSTANCE.Lower(decimalPlaces), FfiConverterTypeRoundingModeINSTANCE.Lower(roundingMode), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)Sqrt() **Decimal {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_sqrt(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Decimal)Sub(other *Decimal) (*Decimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_sub(
		_pointer,FfiConverterDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Decimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Decimal)ToLeBytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*Decimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_decimal_to_le_bytes(
		_pointer, _uniffiStatus)
	}))
}



func (object *Decimal)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterDecimal struct {}

var FfiConverterDecimalINSTANCE = FfiConverterDecimal{}

func (c FfiConverterDecimal) Lift(pointer unsafe.Pointer) *Decimal {
	result := &Decimal {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_decimal(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*Decimal).Destroy)
	return result
}

func (c FfiConverterDecimal) Read(reader io.Reader) *Decimal {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterDecimal) Lower(value *Decimal) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Decimal")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterDecimal) Write(writer io.Writer, value *Decimal) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerDecimal struct {}

func (_ FfiDestroyerDecimal) Destroy(value *Decimal) {
	value.Destroy()
}


type Hash struct {
	ffiObject FfiObject
}
func NewHash(hash []byte) (*Hash, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_hash_new(FfiConverterBytesINSTANCE.Lower(hash), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Hash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func HashFromHexString(hash string) (*Hash, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_hash_from_hex_string(FfiConverterStringINSTANCE.Lower(hash), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Hash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func HashFromUnhashedBytes(bytes []byte) *Hash {
	return FfiConverterHashINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_hash_from_unhashed_bytes(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	}))
}

func HashSborDecode(bytes []byte) (*Hash, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_hash_sbor_decode(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Hash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *Hash)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*Hash")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_hash_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Hash)Bytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*Hash")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_hash_bytes(
		_pointer, _uniffiStatus)
	}))
}



func (object *Hash)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterHash struct {}

var FfiConverterHashINSTANCE = FfiConverterHash{}

func (c FfiConverterHash) Lift(pointer unsafe.Pointer) *Hash {
	result := &Hash {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_hash(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*Hash).Destroy)
	return result
}

func (c FfiConverterHash) Read(reader io.Reader) *Hash {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterHash) Lower(value *Hash) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Hash")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterHash) Write(writer io.Writer, value *Hash) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerHash struct {}

func (_ FfiDestroyerHash) Destroy(value *Hash) {
	value.Destroy()
}


type Instructions struct {
	ffiObject FfiObject
}


func InstructionsFromInstructions(instructions []Instruction, networkId uint8) (*Instructions, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_instructions_from_instructions(FfiConverterSequenceTypeInstructionINSTANCE.Lower(instructions), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Instructions
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterInstructionsINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func InstructionsFromString(string string, networkId uint8) (*Instructions, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_instructions_from_string(FfiConverterStringINSTANCE.Lower(string), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Instructions
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterInstructionsINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *Instructions)AsStr() (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*Instructions")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_instructions_as_str(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Instructions)InstructionsList() []Instruction {
	_pointer := _self.ffiObject.incrementPointer("*Instructions")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypeInstructionINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_instructions_instructions_list(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Instructions)NetworkId() uint8 {
	_pointer := _self.ffiObject.incrementPointer("*Instructions")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint8INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_instructions_network_id(
		_pointer, _uniffiStatus)
	}))
}



func (object *Instructions)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterInstructions struct {}

var FfiConverterInstructionsINSTANCE = FfiConverterInstructions{}

func (c FfiConverterInstructions) Lift(pointer unsafe.Pointer) *Instructions {
	result := &Instructions {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_instructions(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*Instructions).Destroy)
	return result
}

func (c FfiConverterInstructions) Read(reader io.Reader) *Instructions {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterInstructions) Lower(value *Instructions) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Instructions")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterInstructions) Write(writer io.Writer, value *Instructions) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerInstructions struct {}

func (_ FfiDestroyerInstructions) Destroy(value *Instructions) {
	value.Destroy()
}


type Intent struct {
	ffiObject FfiObject
}
func NewIntent(header TransactionHeader, manifest *TransactionManifest, message Message) *Intent {
	return FfiConverterIntentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_intent_new(FfiConverterTypeTransactionHeaderINSTANCE.Lower(header), FfiConverterTransactionManifestINSTANCE.Lower(manifest), FfiConverterTypeMessageINSTANCE.Lower(message), _uniffiStatus)
	}))
}


func IntentDecompile(compiledIntent []byte) (*Intent, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_intent_decompile(FfiConverterBytesINSTANCE.Lower(compiledIntent), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Intent
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterIntentINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *Intent)Compile() ([]byte, error) {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_compile(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Intent)Hash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Intent)Header() TransactionHeader {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeTransactionHeaderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_header(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Intent)IntentHash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_intent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Intent)Manifest() *TransactionManifest {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionManifestINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_manifest(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Intent)Message() Message {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeMessageINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_message(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Intent)StaticallyValidate(validationConfig *ValidationConfig) error {
	_pointer := _self.ffiObject.incrementPointer("*Intent")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_radix_engine_toolkit_uniffi_fn_method_intent_statically_validate(
		_pointer,FfiConverterValidationConfigINSTANCE.Lower(validationConfig), _uniffiStatus)
		return false
	})
		return _uniffiErr
}



func (object *Intent)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterIntent struct {}

var FfiConverterIntentINSTANCE = FfiConverterIntent{}

func (c FfiConverterIntent) Lift(pointer unsafe.Pointer) *Intent {
	result := &Intent {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_intent(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*Intent).Destroy)
	return result
}

func (c FfiConverterIntent) Read(reader io.Reader) *Intent {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterIntent) Lower(value *Intent) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Intent")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterIntent) Write(writer io.Writer, value *Intent) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerIntent struct {}

func (_ FfiDestroyerIntent) Destroy(value *Intent) {
	value.Destroy()
}


type ManifestBuilder struct {
	ffiObject FfiObject
}
func NewManifestBuilder() *ManifestBuilder {
	return FfiConverterManifestBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_manifestbuilder_new( _uniffiStatus)
	}))
}




func (_self *ManifestBuilder)AccessControllerCancelPrimaryRoleBadgeWithdrawAttempt(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_cancel_primary_role_badge_withdraw_attempt(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerCancelPrimaryRoleRecoveryProposal(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_cancel_primary_role_recovery_proposal(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerCancelRecoveryRoleBadgeWithdrawAttempt(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_cancel_recovery_role_badge_withdraw_attempt(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerCancelRecoveryRoleRecoveryProposal(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_cancel_recovery_role_recovery_proposal(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerCreate(controlledAsset ManifestBuilderBucket, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_create(
		_pointer,FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(controlledAsset), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerCreateProof(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_create_proof(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerCreateWithSecurityStructure(controlledAsset ManifestBuilderBucket, primaryRole SecurityStructureRole, recoveryRole SecurityStructureRole, confirmationRole SecurityStructureRole, timedRecoveryDelayInMinutes *uint32, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_create_with_security_structure(
		_pointer,FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(controlledAsset), FfiConverterTypeSecurityStructureRoleINSTANCE.Lower(primaryRole), FfiConverterTypeSecurityStructureRoleINSTANCE.Lower(recoveryRole), FfiConverterTypeSecurityStructureRoleINSTANCE.Lower(confirmationRole), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerInitiateBadgeWithdrawAsPrimary(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_initiate_badge_withdraw_as_primary(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerInitiateBadgeWithdrawAsRecovery(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_initiate_badge_withdraw_as_recovery(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerInitiateRecoveryAsPrimary(address *Address, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_initiate_recovery_as_primary(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerInitiateRecoveryAsRecovery(address *Address, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_initiate_recovery_as_recovery(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerLockPrimaryRole(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_lock_primary_role(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerMintRecoveryBadges(address *Address, nonFungibleLocalIds []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_mint_recovery_badges(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(nonFungibleLocalIds), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerNewFromPublicKeys(controlledAsset ManifestBuilderBucket, primaryRole PublicKey, recoveryRole PublicKey, confirmationRole PublicKey, timedRecoveryDelayInMinutes *uint32, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_new_from_public_keys(
		_pointer,FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(controlledAsset), FfiConverterTypePublicKeyINSTANCE.Lower(primaryRole), FfiConverterTypePublicKeyINSTANCE.Lower(recoveryRole), FfiConverterTypePublicKeyINSTANCE.Lower(confirmationRole), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerQuickConfirmPrimaryRoleBadgeWithdrawAttempt(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_quick_confirm_primary_role_badge_withdraw_attempt(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerQuickConfirmPrimaryRoleRecoveryProposal(address *Address, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_quick_confirm_primary_role_recovery_proposal(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerQuickConfirmRecoveryRoleBadgeWithdrawAttempt(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_quick_confirm_recovery_role_badge_withdraw_attempt(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerQuickConfirmRecoveryRoleRecoveryProposal(address *Address, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_quick_confirm_recovery_role_recovery_proposal(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerStopTimedRecovery(address *Address, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_stop_timed_recovery(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerTimedConfirmRecovery(address *Address, ruleSet RuleSet, timedRecoveryDelayInMinutes *uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_timed_confirm_recovery(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeRuleSetINSTANCE.Lower(ruleSet), FfiConverterOptionalUint32INSTANCE.Lower(timedRecoveryDelayInMinutes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccessControllerUnlockPrimaryRole(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_access_controller_unlock_primary_role(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountAddAuthorizedDepositor(address *Address, badge ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_add_authorized_depositor(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeResourceOrNonFungibleINSTANCE.Lower(badge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountBurn(address *Address, resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_burn(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountBurnNonFungibles(address *Address, resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_burn_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountCreate() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_create(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountCreateAdvanced(ownerRole OwnerRole, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_create_advanced(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountCreateProofOfAmount(address *Address, resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_create_proof_of_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountCreateProofOfNonFungibles(address *Address, resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_create_proof_of_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountDeposit(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_deposit(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountDepositBatch(address *Address, buckets []ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_deposit_batch(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterSequenceTypeManifestBuilderBucketINSTANCE.Lower(buckets), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountDepositEntireWorktop(accountAddress *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_deposit_entire_worktop(
		_pointer,FfiConverterAddressINSTANCE.Lower(accountAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockContingentFee(address *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_lock_contingent_fee(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockFee(address *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_lock_fee(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockFeeAndWithdraw(address *Address, amountToLock *Decimal, resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_lock_fee_and_withdraw(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amountToLock), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockFeeAndWithdrawNonFungibles(address *Address, amountToLock *Decimal, resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_lock_fee_and_withdraw_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amountToLock), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerAirdrop(address *Address, claimants map[string]ResourceSpecifier, bucket ManifestBuilderBucket, tryDirectSend bool) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_airdrop(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterMapStringTypeResourceSpecifierINSTANCE.Lower(claimants), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterBoolINSTANCE.Lower(tryDirectSend), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerClaim(address *Address, claimant *Address, resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_claim(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerClaimNonFungibles(address *Address, claimant *Address, resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_claim_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerGetAmount(address *Address, claimant *Address, resourceAddress *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_get_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterAddressINSTANCE.Lower(resourceAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerGetNonFungibleLocalIds(address *Address, claimant *Address, resourceAddress *Address, limit uint32) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_get_non_fungible_local_ids(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterUint32INSTANCE.Lower(limit), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerInstantiate(ownerRole OwnerRole, storerRole *AccessRule, storerUpdaterRole *AccessRule, recovererRole *AccessRule, recovererUpdaterRole *AccessRule, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_instantiate(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterAccessRuleINSTANCE.Lower(storerRole), FfiConverterAccessRuleINSTANCE.Lower(storerUpdaterRole), FfiConverterAccessRuleINSTANCE.Lower(recovererRole), FfiConverterAccessRuleINSTANCE.Lower(recovererUpdaterRole), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerInstantiateSimple(allowRecover bool) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_instantiate_simple(
		_pointer,FfiConverterBoolINSTANCE.Lower(allowRecover), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerRecover(address *Address, claimant *Address, resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_recover(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerRecoverNonFungibles(address *Address, claimant *Address, resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_recover_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountLockerStore(address *Address, claimant *Address, bucket ManifestBuilderBucket, tryDirectSend bool) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_locker_store(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(claimant), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterBoolINSTANCE.Lower(tryDirectSend), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountRemoveAuthorizedDepositor(address *Address, badge ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_remove_authorized_depositor(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeResourceOrNonFungibleINSTANCE.Lower(badge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountRemoveResourcePreference(address *Address, resourceAddress *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_remove_resource_preference(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountSecurify(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_securify(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountSetDefaultDepositRule(address *Address, defaultDepositRule AccountDefaultDepositRule) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_set_default_deposit_rule(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeAccountDefaultDepositRuleINSTANCE.Lower(defaultDepositRule), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountSetResourcePreference(address *Address, resourceAddress *Address, resourcePreference ResourcePreference) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_set_resource_preference(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterTypeResourcePreferenceINSTANCE.Lower(resourcePreference), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountTryDepositBatchOrAbort(address *Address, buckets []ManifestBuilderBucket, authorizedDepositorBadge *ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_try_deposit_batch_or_abort(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterSequenceTypeManifestBuilderBucketINSTANCE.Lower(buckets), FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE.Lower(authorizedDepositorBadge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountTryDepositBatchOrRefund(address *Address, buckets []ManifestBuilderBucket, authorizedDepositorBadge *ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_try_deposit_batch_or_refund(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterSequenceTypeManifestBuilderBucketINSTANCE.Lower(buckets), FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE.Lower(authorizedDepositorBadge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountTryDepositEntireWorktopOrAbort(accountAddress *Address, authorizedDepositorBadge *ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_try_deposit_entire_worktop_or_abort(
		_pointer,FfiConverterAddressINSTANCE.Lower(accountAddress), FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE.Lower(authorizedDepositorBadge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountTryDepositEntireWorktopOrRefund(accountAddress *Address, authorizedDepositorBadge *ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_try_deposit_entire_worktop_or_refund(
		_pointer,FfiConverterAddressINSTANCE.Lower(accountAddress), FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE.Lower(authorizedDepositorBadge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountTryDepositOrAbort(address *Address, bucket ManifestBuilderBucket, authorizedDepositorBadge *ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_try_deposit_or_abort(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE.Lower(authorizedDepositorBadge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountTryDepositOrRefund(address *Address, bucket ManifestBuilderBucket, authorizedDepositorBadge *ResourceOrNonFungible) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_try_deposit_or_refund(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE.Lower(authorizedDepositorBadge), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountWithdraw(address *Address, resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_withdraw(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AccountWithdrawNonFungibles(address *Address, resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_account_withdraw_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AllocateGlobalAddress(packageAddress *Address, blueprintName string, intoAddressReservation ManifestBuilderAddressReservation, intoNamedAddress ManifestBuilderNamedAddress) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_allocate_global_address(
		_pointer,FfiConverterAddressINSTANCE.Lower(packageAddress), FfiConverterStringINSTANCE.Lower(blueprintName), FfiConverterTypeManifestBuilderAddressReservationINSTANCE.Lower(intoAddressReservation), FfiConverterTypeManifestBuilderNamedAddressINSTANCE.Lower(intoNamedAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AssertWorktopContains(resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_assert_worktop_contains(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AssertWorktopContainsAny(resourceAddress *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_assert_worktop_contains_any(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)AssertWorktopContainsNonFungibles(resourceAddress *Address, ids []NonFungibleLocalId) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_assert_worktop_contains_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)Build(networkId uint8) *TransactionManifest {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionManifestINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_build(
		_pointer,FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	}))
}


func (_self *ManifestBuilder)BurnResource(bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_burn_resource(
		_pointer,FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CallAccessRulesMethod(address ManifestBuilderAddress, methodName string, args []ManifestBuilderValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_call_access_rules_method(
		_pointer,FfiConverterTypeManifestBuilderAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(methodName), FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Lower(args), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CallDirectVaultMethod(address *Address, methodName string, args []ManifestBuilderValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_call_direct_vault_method(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(methodName), FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Lower(args), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CallFunction(address ManifestBuilderAddress, blueprintName string, functionName string, args []ManifestBuilderValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_call_function(
		_pointer,FfiConverterTypeManifestBuilderAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(blueprintName), FfiConverterStringINSTANCE.Lower(functionName), FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Lower(args), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CallMetadataMethod(address ManifestBuilderAddress, methodName string, args []ManifestBuilderValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_call_metadata_method(
		_pointer,FfiConverterTypeManifestBuilderAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(methodName), FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Lower(args), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CallMethod(address ManifestBuilderAddress, methodName string, args []ManifestBuilderValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_call_method(
		_pointer,FfiConverterTypeManifestBuilderAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(methodName), FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Lower(args), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CallRoyaltyMethod(address ManifestBuilderAddress, methodName string, args []ManifestBuilderValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_call_royalty_method(
		_pointer,FfiConverterTypeManifestBuilderAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(methodName), FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Lower(args), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CloneProof(proof ManifestBuilderProof, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_clone_proof(
		_pointer,FfiConverterTypeManifestBuilderProofINSTANCE.Lower(proof), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateFungibleResourceManager(ownerRole OwnerRole, trackTotalSupply bool, divisibility uint8, initialSupply **Decimal, resourceRoles FungibleResourceRoles, metadata MetadataModuleConfig, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_fungible_resource_manager(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterBoolINSTANCE.Lower(trackTotalSupply), FfiConverterUint8INSTANCE.Lower(divisibility), FfiConverterOptionalDecimalINSTANCE.Lower(initialSupply), FfiConverterTypeFungibleResourceRolesINSTANCE.Lower(resourceRoles), FfiConverterTypeMetadataModuleConfigINSTANCE.Lower(metadata), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateProofFromAuthZoneOfAll(resourceAddress *Address, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_proof_from_auth_zone_of_all(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateProofFromAuthZoneOfAmount(resourceAddress *Address, amount *Decimal, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_proof_from_auth_zone_of_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateProofFromAuthZoneOfNonFungibles(resourceAddress *Address, ids []NonFungibleLocalId, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_proof_from_auth_zone_of_non_fungibles(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateProofFromBucketOfAll(bucket ManifestBuilderBucket, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_proof_from_bucket_of_all(
		_pointer,FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateProofFromBucketOfAmount(amount *Decimal, bucket ManifestBuilderBucket, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_proof_from_bucket_of_amount(
		_pointer,FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)CreateProofFromBucketOfNonFungibles(ids []NonFungibleLocalId, bucket ManifestBuilderBucket, intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_create_proof_from_bucket_of_non_fungibles(
		_pointer,FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)DropAllProofs() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_drop_all_proofs(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)DropAuthZoneProofs() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_drop_auth_zone_proofs(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)DropAuthZoneSignatureProofs() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_drop_auth_zone_signature_proofs(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)DropProof(proof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_drop_proof(
		_pointer,FfiConverterTypeManifestBuilderProofINSTANCE.Lower(proof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)FaucetFreeXrd() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_faucet_free_xrd(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)FaucetLockFee() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_faucet_lock_fee(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)IdentityCreate() (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_identity_create(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)IdentityCreateAdvanced(ownerRole OwnerRole) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_identity_create_advanced(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)IdentitySecurify(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_identity_securify(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MetadataGet(address *Address, key string) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_metadata_get(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(key), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MetadataLock(address *Address, key string) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_metadata_lock(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(key), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MetadataRemove(address *Address, key string) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_metadata_remove(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(key), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MetadataSet(address *Address, key string, value MetadataValue) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_metadata_set(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(key), FfiConverterTypeMetadataValueINSTANCE.Lower(value), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MintFungible(resourceAddress *Address, amount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_mint_fungible(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolContribute(address *Address, buckets []ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_contribute(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterSequenceTypeManifestBuilderBucketINSTANCE.Lower(buckets), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolGetRedemptionValue(address *Address, amountOfPoolUnits *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_get_redemption_value(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amountOfPoolUnits), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolGetVaultAmount(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_get_vault_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolInstantiate(ownerRole OwnerRole, poolManagerRule *AccessRule, resourceAddresses []*Address, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_instantiate(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterAccessRuleINSTANCE.Lower(poolManagerRule), FfiConverterSequenceAddressINSTANCE.Lower(resourceAddresses), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolProtectedDeposit(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_protected_deposit(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolProtectedWithdraw(address *Address, resourceAddress *Address, amount *Decimal, withdrawStrategy WithdrawStrategy) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_protected_withdraw(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterTypeWithdrawStrategyINSTANCE.Lower(withdrawStrategy), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)MultiResourcePoolRedeem(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_multi_resource_pool_redeem(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolContribute(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_contribute(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolGetRedemptionValue(address *Address, amountOfPoolUnits *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_get_redemption_value(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amountOfPoolUnits), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolGetVaultAmount(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_get_vault_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolInstantiate(ownerRole OwnerRole, poolManagerRule *AccessRule, resourceAddress *Address, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_instantiate(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterAccessRuleINSTANCE.Lower(poolManagerRule), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolProtectedDeposit(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_protected_deposit(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolProtectedWithdraw(address *Address, amount *Decimal, withdrawStrategy WithdrawStrategy) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_protected_withdraw(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterTypeWithdrawStrategyINSTANCE.Lower(withdrawStrategy), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)OneResourcePoolRedeem(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_one_resource_pool_redeem(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)PackageClaimRoyalty(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_package_claim_royalty(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)PackagePublish(code []byte, definition []byte, metadata map[string]MetadataInitEntry) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_package_publish(
		_pointer,FfiConverterBytesINSTANCE.Lower(code), FfiConverterBytesINSTANCE.Lower(definition), FfiConverterMapStringTypeMetadataInitEntryINSTANCE.Lower(metadata), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)PackagePublishAdvanced(ownerRole OwnerRole, code []byte, definition []byte, metadata map[string]MetadataInitEntry, packageAddress *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_package_publish_advanced(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterBytesINSTANCE.Lower(code), FfiConverterBytesINSTANCE.Lower(definition), FfiConverterMapStringTypeMetadataInitEntryINSTANCE.Lower(metadata), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(packageAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)PopFromAuthZone(intoProof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_pop_from_auth_zone(
		_pointer,FfiConverterTypeManifestBuilderProofINSTANCE.Lower(intoProof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)PushToAuthZone(proof ManifestBuilderProof) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_push_to_auth_zone(
		_pointer,FfiConverterTypeManifestBuilderProofINSTANCE.Lower(proof), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ReturnToWorktop(bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_return_to_worktop(
		_pointer,FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoleAssignmentGet(address *Address, module ModuleId, roleKey string) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_role_assignment_get(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeModuleIdINSTANCE.Lower(module), FfiConverterStringINSTANCE.Lower(roleKey), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoleAssignmentLockOwner(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_role_assignment_lock_owner(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoleAssignmentSet(address *Address, module ModuleId, roleKey string, rule *AccessRule) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_role_assignment_set(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeModuleIdINSTANCE.Lower(module), FfiConverterStringINSTANCE.Lower(roleKey), FfiConverterAccessRuleINSTANCE.Lower(rule), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoleAssignmentSetOwner(address *Address, rule *AccessRule) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_role_assignment_set_owner(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAccessRuleINSTANCE.Lower(rule), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoyaltyClaim(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_royalty_claim(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoyaltyLock(address *Address, method string) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_royalty_lock(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(method), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)RoyaltySet(address *Address, method string, amount RoyaltyAmount) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_royalty_set(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(method), FfiConverterTypeRoyaltyAmountINSTANCE.Lower(amount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TakeAllFromWorktop(resourceAddress *Address, intoBucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_take_all_from_worktop(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(intoBucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TakeFromWorktop(resourceAddress *Address, amount *Decimal, intoBucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_take_from_worktop(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(intoBucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TakeNonFungiblesFromWorktop(resourceAddress *Address, ids []NonFungibleLocalId, intoBucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_take_non_fungibles_from_worktop(
		_pointer,FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Lower(ids), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(intoBucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolContribute(address *Address, buckets []ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_contribute(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterSequenceTypeManifestBuilderBucketINSTANCE.Lower(buckets), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolGetRedemptionValue(address *Address, amountOfPoolUnits *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_get_redemption_value(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amountOfPoolUnits), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolGetVaultAmount(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_get_vault_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolInstantiate(ownerRole OwnerRole, poolManagerRule *AccessRule, resourceAddresses []*Address, addressReservation *ManifestBuilderAddressReservation) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_instantiate(
		_pointer,FfiConverterTypeOwnerRoleINSTANCE.Lower(ownerRole), FfiConverterAccessRuleINSTANCE.Lower(poolManagerRule), FfiConverterSequenceAddressINSTANCE.Lower(resourceAddresses), FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE.Lower(addressReservation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolProtectedDeposit(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_protected_deposit(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolProtectedWithdraw(address *Address, resourceAddress *Address, amount *Decimal, withdrawStrategy WithdrawStrategy) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_protected_withdraw(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterDecimalINSTANCE.Lower(amount), FfiConverterTypeWithdrawStrategyINSTANCE.Lower(withdrawStrategy), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)TwoResourcePoolRedeem(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_two_resource_pool_redeem(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorAcceptsDelegatedStake(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_accepts_delegated_stake(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorClaimXrd(address *Address, bucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_claim_xrd(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(bucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorFinishUnlockOwnerStakeUnits(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_finish_unlock_owner_stake_units(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorGetProtocolUpdateReadiness(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_get_protocol_update_readiness(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorGetRedemptionValue(address *Address, amountOfStakeUnits *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_get_redemption_value(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(amountOfStakeUnits), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorLockOwnerStakeUnits(address *Address, stakeUnitBucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_lock_owner_stake_units(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(stakeUnitBucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorRegister(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_register(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorSignalProtocolUpdateReadiness(address *Address, vote string) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_signal_protocol_update_readiness(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(vote), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorStake(address *Address, stake ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_stake(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(stake), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorStakeAsOwner(address *Address, stake ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_stake_as_owner(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(stake), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorStartUnlockOwnerStakeUnits(address *Address, requestedStakeUnitAmount *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_start_unlock_owner_stake_units(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(requestedStakeUnitAmount), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorTotalStakeUnitSupply(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_total_stake_unit_supply(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorTotalStakeXrdAmount(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_total_stake_xrd_amount(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorUnregister(address *Address) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_unregister(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorUnstake(address *Address, stakeUnitBucket ManifestBuilderBucket) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_unstake(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypeManifestBuilderBucketINSTANCE.Lower(stakeUnitBucket), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorUpdateAcceptDelegatedStake(address *Address, acceptDelegatedStake bool) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_update_accept_delegated_stake(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterBoolINSTANCE.Lower(acceptDelegatedStake), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorUpdateFee(address *Address, newFeeFactor *Decimal) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_update_fee(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterDecimalINSTANCE.Lower(newFeeFactor), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *ManifestBuilder)ValidatorUpdateKey(address *Address, key PublicKey) (*ManifestBuilder, error) {
	_pointer := _self.ffiObject.incrementPointer("*ManifestBuilder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_manifestbuilder_validator_update_key(
		_pointer,FfiConverterAddressINSTANCE.Lower(address), FfiConverterTypePublicKeyINSTANCE.Lower(key), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *ManifestBuilder
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterManifestBuilderINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (object *ManifestBuilder)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterManifestBuilder struct {}

var FfiConverterManifestBuilderINSTANCE = FfiConverterManifestBuilder{}

func (c FfiConverterManifestBuilder) Lift(pointer unsafe.Pointer) *ManifestBuilder {
	result := &ManifestBuilder {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_manifestbuilder(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*ManifestBuilder).Destroy)
	return result
}

func (c FfiConverterManifestBuilder) Read(reader io.Reader) *ManifestBuilder {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterManifestBuilder) Lower(value *ManifestBuilder) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*ManifestBuilder")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterManifestBuilder) Write(writer io.Writer, value *ManifestBuilder) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerManifestBuilder struct {}

func (_ FfiDestroyerManifestBuilder) Destroy(value *ManifestBuilder) {
	value.Destroy()
}


type MessageValidationConfig struct {
	ffiObject FfiObject
}
func NewMessageValidationConfig(maxPlaintextMessageLength uint64, maxEncryptedMessageLength uint64, maxMimeTypeLength uint64, maxDecryptors uint64) *MessageValidationConfig {
	return FfiConverterMessageValidationConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_messagevalidationconfig_new(FfiConverterUint64INSTANCE.Lower(maxPlaintextMessageLength), FfiConverterUint64INSTANCE.Lower(maxEncryptedMessageLength), FfiConverterUint64INSTANCE.Lower(maxMimeTypeLength), FfiConverterUint64INSTANCE.Lower(maxDecryptors), _uniffiStatus)
	}))
}


func MessageValidationConfigDefault() *MessageValidationConfig {
	return FfiConverterMessageValidationConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_messagevalidationconfig_default( _uniffiStatus)
	}))
}



func (_self *MessageValidationConfig)MaxDecryptors() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*MessageValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_messagevalidationconfig_max_decryptors(
		_pointer, _uniffiStatus)
	}))
}


func (_self *MessageValidationConfig)MaxEncryptedMessageLength() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*MessageValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_messagevalidationconfig_max_encrypted_message_length(
		_pointer, _uniffiStatus)
	}))
}


func (_self *MessageValidationConfig)MaxMimeTypeLength() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*MessageValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_messagevalidationconfig_max_mime_type_length(
		_pointer, _uniffiStatus)
	}))
}


func (_self *MessageValidationConfig)MaxPlaintextMessageLength() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*MessageValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_messagevalidationconfig_max_plaintext_message_length(
		_pointer, _uniffiStatus)
	}))
}



func (object *MessageValidationConfig)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterMessageValidationConfig struct {}

var FfiConverterMessageValidationConfigINSTANCE = FfiConverterMessageValidationConfig{}

func (c FfiConverterMessageValidationConfig) Lift(pointer unsafe.Pointer) *MessageValidationConfig {
	result := &MessageValidationConfig {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_messagevalidationconfig(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*MessageValidationConfig).Destroy)
	return result
}

func (c FfiConverterMessageValidationConfig) Read(reader io.Reader) *MessageValidationConfig {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterMessageValidationConfig) Lower(value *MessageValidationConfig) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*MessageValidationConfig")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterMessageValidationConfig) Write(writer io.Writer, value *MessageValidationConfig) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerMessageValidationConfig struct {}

func (_ FfiDestroyerMessageValidationConfig) Destroy(value *MessageValidationConfig) {
	value.Destroy()
}


type NonFungibleGlobalId struct {
	ffiObject FfiObject
}
func NewNonFungibleGlobalId(nonFungibleGlobalId string) (*NonFungibleGlobalId, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_nonfungibleglobalid_new(FfiConverterStringINSTANCE.Lower(nonFungibleGlobalId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NonFungibleGlobalId
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNonFungibleGlobalIdINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func NonFungibleGlobalIdFromParts(resourceAddress *Address, nonFungibleLocalId NonFungibleLocalId) (*NonFungibleGlobalId, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_nonfungibleglobalid_from_parts(FfiConverterAddressINSTANCE.Lower(resourceAddress), FfiConverterTypeNonFungibleLocalIdINSTANCE.Lower(nonFungibleLocalId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NonFungibleGlobalId
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNonFungibleGlobalIdINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func NonFungibleGlobalIdVirtualSignatureBadge(publicKey PublicKey, networkId uint8) (*NonFungibleGlobalId, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_nonfungibleglobalid_virtual_signature_badge(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NonFungibleGlobalId
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNonFungibleGlobalIdINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *NonFungibleGlobalId)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*NonFungibleGlobalId")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_nonfungibleglobalid_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *NonFungibleGlobalId)LocalId() NonFungibleLocalId {
	_pointer := _self.ffiObject.incrementPointer("*NonFungibleGlobalId")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeNonFungibleLocalIdINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_nonfungibleglobalid_local_id(
		_pointer, _uniffiStatus)
	}))
}


func (_self *NonFungibleGlobalId)ResourceAddress() *Address {
	_pointer := _self.ffiObject.incrementPointer("*NonFungibleGlobalId")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterAddressINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_nonfungibleglobalid_resource_address(
		_pointer, _uniffiStatus)
	}))
}



func (object *NonFungibleGlobalId)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterNonFungibleGlobalId struct {}

var FfiConverterNonFungibleGlobalIdINSTANCE = FfiConverterNonFungibleGlobalId{}

func (c FfiConverterNonFungibleGlobalId) Lift(pointer unsafe.Pointer) *NonFungibleGlobalId {
	result := &NonFungibleGlobalId {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_nonfungibleglobalid(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*NonFungibleGlobalId).Destroy)
	return result
}

func (c FfiConverterNonFungibleGlobalId) Read(reader io.Reader) *NonFungibleGlobalId {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterNonFungibleGlobalId) Lower(value *NonFungibleGlobalId) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*NonFungibleGlobalId")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterNonFungibleGlobalId) Write(writer io.Writer, value *NonFungibleGlobalId) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerNonFungibleGlobalId struct {}

func (_ FfiDestroyerNonFungibleGlobalId) Destroy(value *NonFungibleGlobalId) {
	value.Destroy()
}


type NotarizedTransaction struct {
	ffiObject FfiObject
}
func NewNotarizedTransaction(signedIntent *SignedIntent, notarySignature Signature) *NotarizedTransaction {
	return FfiConverterNotarizedTransactionINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_notarizedtransaction_new(FfiConverterSignedIntentINSTANCE.Lower(signedIntent), FfiConverterTypeSignatureINSTANCE.Lower(notarySignature), _uniffiStatus)
	}))
}


func NotarizedTransactionDecompile(compiledNotarizedTransaction []byte) (*NotarizedTransaction, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_notarizedtransaction_decompile(FfiConverterBytesINSTANCE.Lower(compiledNotarizedTransaction), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NotarizedTransaction
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNotarizedTransactionINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *NotarizedTransaction)Compile() ([]byte, error) {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_compile(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *NotarizedTransaction)Hash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *NotarizedTransaction)IntentHash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_intent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *NotarizedTransaction)NotarizedTransactionHash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_notarized_transaction_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *NotarizedTransaction)NotarySignature() Signature {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeSignatureINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_notary_signature(
		_pointer, _uniffiStatus)
	}))
}


func (_self *NotarizedTransaction)SignedIntent() *SignedIntent {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSignedIntentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_signed_intent(
		_pointer, _uniffiStatus)
	}))
}


func (_self *NotarizedTransaction)SignedIntentHash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_signed_intent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *NotarizedTransaction)StaticallyValidate(validationConfig *ValidationConfig) error {
	_pointer := _self.ffiObject.incrementPointer("*NotarizedTransaction")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_radix_engine_toolkit_uniffi_fn_method_notarizedtransaction_statically_validate(
		_pointer,FfiConverterValidationConfigINSTANCE.Lower(validationConfig), _uniffiStatus)
		return false
	})
		return _uniffiErr
}



func (object *NotarizedTransaction)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterNotarizedTransaction struct {}

var FfiConverterNotarizedTransactionINSTANCE = FfiConverterNotarizedTransaction{}

func (c FfiConverterNotarizedTransaction) Lift(pointer unsafe.Pointer) *NotarizedTransaction {
	result := &NotarizedTransaction {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_notarizedtransaction(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*NotarizedTransaction).Destroy)
	return result
}

func (c FfiConverterNotarizedTransaction) Read(reader io.Reader) *NotarizedTransaction {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterNotarizedTransaction) Lower(value *NotarizedTransaction) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*NotarizedTransaction")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterNotarizedTransaction) Write(writer io.Writer, value *NotarizedTransaction) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerNotarizedTransaction struct {}

func (_ FfiDestroyerNotarizedTransaction) Destroy(value *NotarizedTransaction) {
	value.Destroy()
}


type OlympiaAddress struct {
	ffiObject FfiObject
}
func NewOlympiaAddress(address string) *OlympiaAddress {
	return FfiConverterOlympiaAddressINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_olympiaaddress_new(FfiConverterStringINSTANCE.Lower(address), _uniffiStatus)
	}))
}




func (_self *OlympiaAddress)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*OlympiaAddress")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_olympiaaddress_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *OlympiaAddress)PublicKey() (PublicKey, error) {
	_pointer := _self.ffiObject.incrementPointer("*OlympiaAddress")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_olympiaaddress_public_key(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue PublicKey
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypePublicKeyINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (object *OlympiaAddress)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOlympiaAddress struct {}

var FfiConverterOlympiaAddressINSTANCE = FfiConverterOlympiaAddress{}

func (c FfiConverterOlympiaAddress) Lift(pointer unsafe.Pointer) *OlympiaAddress {
	result := &OlympiaAddress {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_olympiaaddress(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*OlympiaAddress).Destroy)
	return result
}

func (c FfiConverterOlympiaAddress) Read(reader io.Reader) *OlympiaAddress {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOlympiaAddress) Lower(value *OlympiaAddress) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OlympiaAddress")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterOlympiaAddress) Write(writer io.Writer, value *OlympiaAddress) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOlympiaAddress struct {}

func (_ FfiDestroyerOlympiaAddress) Destroy(value *OlympiaAddress) {
	value.Destroy()
}


type PreciseDecimal struct {
	ffiObject FfiObject
}
func NewPreciseDecimal(value string) (*PreciseDecimal, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_precisedecimal_new(FfiConverterStringINSTANCE.Lower(value), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func PreciseDecimalFromLeBytes(value []byte) *PreciseDecimal {
	return FfiConverterPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_precisedecimal_from_le_bytes(FfiConverterBytesINSTANCE.Lower(value), _uniffiStatus)
	}))
}

func PreciseDecimalMax() *PreciseDecimal {
	return FfiConverterPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_precisedecimal_max( _uniffiStatus)
	}))
}

func PreciseDecimalMin() *PreciseDecimal {
	return FfiConverterPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_precisedecimal_min( _uniffiStatus)
	}))
}

func PreciseDecimalOne() *PreciseDecimal {
	return FfiConverterPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_precisedecimal_one( _uniffiStatus)
	}))
}

func PreciseDecimalZero() *PreciseDecimal {
	return FfiConverterPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_precisedecimal_zero( _uniffiStatus)
	}))
}



func (_self *PreciseDecimal)Abs() (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_abs(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)Add(other *PreciseDecimal) (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_add(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)Cbrt() (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_cbrt(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)Ceiling() (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_ceiling(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)Div(other *PreciseDecimal) (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_div(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)Equal(other *PreciseDecimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_equal(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)Floor() (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_floor(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)GreaterThan(other *PreciseDecimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_greater_than(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)GreaterThanOrEqual(other *PreciseDecimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_greater_than_or_equal(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)IsNegative() bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_is_negative(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)IsPositive() bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_is_positive(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)IsZero() bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_is_zero(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)LessThan(other *PreciseDecimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_less_than(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)LessThanOrEqual(other *PreciseDecimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_less_than_or_equal(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)Mantissa() string {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_mantissa(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)Mul(other *PreciseDecimal) (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_mul(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)NotEqual(other *PreciseDecimal) bool {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_not_equal(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)NthRoot(n uint32) **PreciseDecimal {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_nth_root(
		_pointer,FfiConverterUint32INSTANCE.Lower(n), _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)Powi(exp int64) (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_powi(
		_pointer,FfiConverterInt64INSTANCE.Lower(exp), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)Round(decimalPlaces int32, roundingMode RoundingMode) (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_round(
		_pointer,FfiConverterInt32INSTANCE.Lower(decimalPlaces), FfiConverterTypeRoundingModeINSTANCE.Lower(roundingMode), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)Sqrt() **PreciseDecimal {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalPreciseDecimalINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_sqrt(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PreciseDecimal)Sub(other *PreciseDecimal) (*PreciseDecimal, error) {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_sub(
		_pointer,FfiConverterPreciseDecimalINSTANCE.Lower(other), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PreciseDecimal
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPreciseDecimalINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *PreciseDecimal)ToLeBytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*PreciseDecimal")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_precisedecimal_to_le_bytes(
		_pointer, _uniffiStatus)
	}))
}



func (object *PreciseDecimal)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterPreciseDecimal struct {}

var FfiConverterPreciseDecimalINSTANCE = FfiConverterPreciseDecimal{}

func (c FfiConverterPreciseDecimal) Lift(pointer unsafe.Pointer) *PreciseDecimal {
	result := &PreciseDecimal {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_precisedecimal(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*PreciseDecimal).Destroy)
	return result
}

func (c FfiConverterPreciseDecimal) Read(reader io.Reader) *PreciseDecimal {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterPreciseDecimal) Lower(value *PreciseDecimal) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*PreciseDecimal")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterPreciseDecimal) Write(writer io.Writer, value *PreciseDecimal) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerPreciseDecimal struct {}

func (_ FfiDestroyerPreciseDecimal) Destroy(value *PreciseDecimal) {
	value.Destroy()
}


type PrivateKey struct {
	ffiObject FfiObject
}
func NewPrivateKey(bytes []byte, curve Curve) (*PrivateKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_privatekey_new(FfiConverterBytesINSTANCE.Lower(bytes), FfiConverterTypeCurveINSTANCE.Lower(curve), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PrivateKey
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPrivateKeyINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func PrivateKeyNewEd25519(bytes []byte) (*PrivateKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_privatekey_new_ed25519(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PrivateKey
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPrivateKeyINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func PrivateKeyNewSecp256k1(bytes []byte) (*PrivateKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_privatekey_new_secp256k1(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *PrivateKey
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterPrivateKeyINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *PrivateKey)Curve() Curve {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeCurveINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_curve(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PrivateKey)PublicKey() PublicKey {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypePublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_public_key(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PrivateKey)PublicKeyBytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_public_key_bytes(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PrivateKey)Raw() []byte {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_raw(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PrivateKey)RawHex() string {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_raw_hex(
		_pointer, _uniffiStatus)
	}))
}


func (_self *PrivateKey)Sign(hash *Hash) []byte {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_sign(
		_pointer,FfiConverterHashINSTANCE.Lower(hash), _uniffiStatus)
	}))
}


func (_self *PrivateKey)SignToSignature(hash *Hash) Signature {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeSignatureINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_sign_to_signature(
		_pointer,FfiConverterHashINSTANCE.Lower(hash), _uniffiStatus)
	}))
}


func (_self *PrivateKey)SignToSignatureWithPublicKey(hash *Hash) SignatureWithPublicKey {
	_pointer := _self.ffiObject.incrementPointer("*PrivateKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeSignatureWithPublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_privatekey_sign_to_signature_with_public_key(
		_pointer,FfiConverterHashINSTANCE.Lower(hash), _uniffiStatus)
	}))
}



func (object *PrivateKey)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterPrivateKey struct {}

var FfiConverterPrivateKeyINSTANCE = FfiConverterPrivateKey{}

func (c FfiConverterPrivateKey) Lift(pointer unsafe.Pointer) *PrivateKey {
	result := &PrivateKey {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_privatekey(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*PrivateKey).Destroy)
	return result
}

func (c FfiConverterPrivateKey) Read(reader io.Reader) *PrivateKey {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterPrivateKey) Lower(value *PrivateKey) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*PrivateKey")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterPrivateKey) Write(writer io.Writer, value *PrivateKey) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerPrivateKey struct {}

func (_ FfiDestroyerPrivateKey) Destroy(value *PrivateKey) {
	value.Destroy()
}


type SignedIntent struct {
	ffiObject FfiObject
}
func NewSignedIntent(intent *Intent, intentSignatures []SignatureWithPublicKey) *SignedIntent {
	return FfiConverterSignedIntentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_signedintent_new(FfiConverterIntentINSTANCE.Lower(intent), FfiConverterSequenceTypeSignatureWithPublicKeyINSTANCE.Lower(intentSignatures), _uniffiStatus)
	}))
}


func SignedIntentDecompile(compiledSignedIntent []byte) (*SignedIntent, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_signedintent_decompile(FfiConverterBytesINSTANCE.Lower(compiledSignedIntent), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *SignedIntent
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterSignedIntentINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *SignedIntent)Compile() ([]byte, error) {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_compile(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *SignedIntent)Hash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *SignedIntent)Intent() *Intent {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterIntentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_intent(
		_pointer, _uniffiStatus)
	}))
}


func (_self *SignedIntent)IntentHash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_intent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *SignedIntent)IntentSignatures() []SignatureWithPublicKey {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypeSignatureWithPublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_intent_signatures(
		_pointer, _uniffiStatus)
	}))
}


func (_self *SignedIntent)SignedIntentHash() (*TransactionHash, error) {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_signed_intent_hash(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *SignedIntent)StaticallyValidate(validationConfig *ValidationConfig) error {
	_pointer := _self.ffiObject.incrementPointer("*SignedIntent")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_radix_engine_toolkit_uniffi_fn_method_signedintent_statically_validate(
		_pointer,FfiConverterValidationConfigINSTANCE.Lower(validationConfig), _uniffiStatus)
		return false
	})
		return _uniffiErr
}



func (object *SignedIntent)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterSignedIntent struct {}

var FfiConverterSignedIntentINSTANCE = FfiConverterSignedIntent{}

func (c FfiConverterSignedIntent) Lift(pointer unsafe.Pointer) *SignedIntent {
	result := &SignedIntent {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_signedintent(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*SignedIntent).Destroy)
	return result
}

func (c FfiConverterSignedIntent) Read(reader io.Reader) *SignedIntent {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterSignedIntent) Lower(value *SignedIntent) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*SignedIntent")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterSignedIntent) Write(writer io.Writer, value *SignedIntent) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerSignedIntent struct {}

func (_ FfiDestroyerSignedIntent) Destroy(value *SignedIntent) {
	value.Destroy()
}


type TransactionBuilder struct {
	ffiObject FfiObject
}
func NewTransactionBuilder() *TransactionBuilder {
	return FfiConverterTransactionBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_transactionbuilder_new( _uniffiStatus)
	}))
}




func (_self *TransactionBuilder)Header(header TransactionHeader) *TransactionBuilderHeaderStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderHeaderStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuilder_header(
		_pointer,FfiConverterTypeTransactionHeaderINSTANCE.Lower(header), _uniffiStatus)
	}))
}



func (object *TransactionBuilder)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTransactionBuilder struct {}

var FfiConverterTransactionBuilderINSTANCE = FfiConverterTransactionBuilder{}

func (c FfiConverterTransactionBuilder) Lift(pointer unsafe.Pointer) *TransactionBuilder {
	result := &TransactionBuilder {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_transactionbuilder(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*TransactionBuilder).Destroy)
	return result
}

func (c FfiConverterTransactionBuilder) Read(reader io.Reader) *TransactionBuilder {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTransactionBuilder) Lower(value *TransactionBuilder) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TransactionBuilder")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterTransactionBuilder) Write(writer io.Writer, value *TransactionBuilder) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTransactionBuilder struct {}

func (_ FfiDestroyerTransactionBuilder) Destroy(value *TransactionBuilder) {
	value.Destroy()
}


type TransactionBuilderHeaderStep struct {
	ffiObject FfiObject
}




func (_self *TransactionBuilderHeaderStep)Manifest(manifest *TransactionManifest) *TransactionBuilderMessageStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderHeaderStep")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderMessageStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuilderheaderstep_manifest(
		_pointer,FfiConverterTransactionManifestINSTANCE.Lower(manifest), _uniffiStatus)
	}))
}



func (object *TransactionBuilderHeaderStep)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTransactionBuilderHeaderStep struct {}

var FfiConverterTransactionBuilderHeaderStepINSTANCE = FfiConverterTransactionBuilderHeaderStep{}

func (c FfiConverterTransactionBuilderHeaderStep) Lift(pointer unsafe.Pointer) *TransactionBuilderHeaderStep {
	result := &TransactionBuilderHeaderStep {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_transactionbuilderheaderstep(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*TransactionBuilderHeaderStep).Destroy)
	return result
}

func (c FfiConverterTransactionBuilderHeaderStep) Read(reader io.Reader) *TransactionBuilderHeaderStep {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTransactionBuilderHeaderStep) Lower(value *TransactionBuilderHeaderStep) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TransactionBuilderHeaderStep")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterTransactionBuilderHeaderStep) Write(writer io.Writer, value *TransactionBuilderHeaderStep) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTransactionBuilderHeaderStep struct {}

func (_ FfiDestroyerTransactionBuilderHeaderStep) Destroy(value *TransactionBuilderHeaderStep) {
	value.Destroy()
}


type TransactionBuilderIntentSignaturesStep struct {
	ffiObject FfiObject
}
func NewTransactionBuilderIntentSignaturesStep(messageStep *TransactionBuilderMessageStep) *TransactionBuilderIntentSignaturesStep {
	return FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_transactionbuilderintentsignaturesstep_new(FfiConverterTransactionBuilderMessageStepINSTANCE.Lower(messageStep), _uniffiStatus)
	}))
}




func (_self *TransactionBuilderIntentSignaturesStep)NotarizeWithPrivateKey(privateKey *PrivateKey) (*NotarizedTransaction, error) {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderIntentSignaturesStep")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuilderintentsignaturesstep_notarize_with_private_key(
		_pointer,FfiConverterPrivateKeyINSTANCE.Lower(privateKey), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NotarizedTransaction
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNotarizedTransactionINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *TransactionBuilderIntentSignaturesStep)NotarizeWithSigner(signer Signer) (*NotarizedTransaction, error) {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderIntentSignaturesStep")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuilderintentsignaturesstep_notarize_with_signer(
		_pointer,FfiConverterCallbackInterfaceSignerINSTANCE.Lower(signer), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NotarizedTransaction
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNotarizedTransactionINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *TransactionBuilderIntentSignaturesStep)SignWithPrivateKey(privateKey *PrivateKey) *TransactionBuilderIntentSignaturesStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderIntentSignaturesStep")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuilderintentsignaturesstep_sign_with_private_key(
		_pointer,FfiConverterPrivateKeyINSTANCE.Lower(privateKey), _uniffiStatus)
	}))
}


func (_self *TransactionBuilderIntentSignaturesStep)SignWithSigner(signer Signer) *TransactionBuilderIntentSignaturesStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderIntentSignaturesStep")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuilderintentsignaturesstep_sign_with_signer(
		_pointer,FfiConverterCallbackInterfaceSignerINSTANCE.Lower(signer), _uniffiStatus)
	}))
}



func (object *TransactionBuilderIntentSignaturesStep)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTransactionBuilderIntentSignaturesStep struct {}

var FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE = FfiConverterTransactionBuilderIntentSignaturesStep{}

func (c FfiConverterTransactionBuilderIntentSignaturesStep) Lift(pointer unsafe.Pointer) *TransactionBuilderIntentSignaturesStep {
	result := &TransactionBuilderIntentSignaturesStep {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_transactionbuilderintentsignaturesstep(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*TransactionBuilderIntentSignaturesStep).Destroy)
	return result
}

func (c FfiConverterTransactionBuilderIntentSignaturesStep) Read(reader io.Reader) *TransactionBuilderIntentSignaturesStep {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTransactionBuilderIntentSignaturesStep) Lower(value *TransactionBuilderIntentSignaturesStep) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TransactionBuilderIntentSignaturesStep")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterTransactionBuilderIntentSignaturesStep) Write(writer io.Writer, value *TransactionBuilderIntentSignaturesStep) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTransactionBuilderIntentSignaturesStep struct {}

func (_ FfiDestroyerTransactionBuilderIntentSignaturesStep) Destroy(value *TransactionBuilderIntentSignaturesStep) {
	value.Destroy()
}


type TransactionBuilderMessageStep struct {
	ffiObject FfiObject
}




func (_self *TransactionBuilderMessageStep)Message(message Message) *TransactionBuilderIntentSignaturesStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderMessageStep")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuildermessagestep_message(
		_pointer,FfiConverterTypeMessageINSTANCE.Lower(message), _uniffiStatus)
	}))
}


func (_self *TransactionBuilderMessageStep)SignWithPrivateKey(privateKey *PrivateKey) *TransactionBuilderIntentSignaturesStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderMessageStep")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuildermessagestep_sign_with_private_key(
		_pointer,FfiConverterPrivateKeyINSTANCE.Lower(privateKey), _uniffiStatus)
	}))
}


func (_self *TransactionBuilderMessageStep)SignWithSigner(signer Signer) *TransactionBuilderIntentSignaturesStep {
	_pointer := _self.ffiObject.incrementPointer("*TransactionBuilderMessageStep")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTransactionBuilderIntentSignaturesStepINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionbuildermessagestep_sign_with_signer(
		_pointer,FfiConverterCallbackInterfaceSignerINSTANCE.Lower(signer), _uniffiStatus)
	}))
}



func (object *TransactionBuilderMessageStep)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTransactionBuilderMessageStep struct {}

var FfiConverterTransactionBuilderMessageStepINSTANCE = FfiConverterTransactionBuilderMessageStep{}

func (c FfiConverterTransactionBuilderMessageStep) Lift(pointer unsafe.Pointer) *TransactionBuilderMessageStep {
	result := &TransactionBuilderMessageStep {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_transactionbuildermessagestep(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*TransactionBuilderMessageStep).Destroy)
	return result
}

func (c FfiConverterTransactionBuilderMessageStep) Read(reader io.Reader) *TransactionBuilderMessageStep {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTransactionBuilderMessageStep) Lower(value *TransactionBuilderMessageStep) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TransactionBuilderMessageStep")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterTransactionBuilderMessageStep) Write(writer io.Writer, value *TransactionBuilderMessageStep) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTransactionBuilderMessageStep struct {}

func (_ FfiDestroyerTransactionBuilderMessageStep) Destroy(value *TransactionBuilderMessageStep) {
	value.Destroy()
}


type TransactionHash struct {
	ffiObject FfiObject
}


func TransactionHashFromStr(string string, networkId uint8) (*TransactionHash, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_transactionhash_from_str(FfiConverterStringINSTANCE.Lower(string), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionHash
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionHashINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *TransactionHash)AsHash() *Hash {
	_pointer := _self.ffiObject.incrementPointer("*TransactionHash")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterHashINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionhash_as_hash(
		_pointer, _uniffiStatus)
	}))
}


func (_self *TransactionHash)AsStr() string {
	_pointer := _self.ffiObject.incrementPointer("*TransactionHash")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionhash_as_str(
		_pointer, _uniffiStatus)
	}))
}


func (_self *TransactionHash)Bytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*TransactionHash")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionhash_bytes(
		_pointer, _uniffiStatus)
	}))
}


func (_self *TransactionHash)NetworkId() uint8 {
	_pointer := _self.ffiObject.incrementPointer("*TransactionHash")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint8INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionhash_network_id(
		_pointer, _uniffiStatus)
	}))
}



func (object *TransactionHash)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTransactionHash struct {}

var FfiConverterTransactionHashINSTANCE = FfiConverterTransactionHash{}

func (c FfiConverterTransactionHash) Lift(pointer unsafe.Pointer) *TransactionHash {
	result := &TransactionHash {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_transactionhash(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*TransactionHash).Destroy)
	return result
}

func (c FfiConverterTransactionHash) Read(reader io.Reader) *TransactionHash {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTransactionHash) Lower(value *TransactionHash) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TransactionHash")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterTransactionHash) Write(writer io.Writer, value *TransactionHash) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTransactionHash struct {}

func (_ FfiDestroyerTransactionHash) Destroy(value *TransactionHash) {
	value.Destroy()
}


type TransactionManifest struct {
	ffiObject FfiObject
}
func NewTransactionManifest(instructions *Instructions, blobs [][]byte) *TransactionManifest {
	return FfiConverterTransactionManifestINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_transactionmanifest_new(FfiConverterInstructionsINSTANCE.Lower(instructions), FfiConverterSequenceBytesINSTANCE.Lower(blobs), _uniffiStatus)
	}))
}


func TransactionManifestDecompile(compiled []byte, networkId uint8) (*TransactionManifest, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_transactionmanifest_decompile(FfiConverterBytesINSTANCE.Lower(compiled), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionManifest
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionManifestINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (_self *TransactionManifest)Blobs() [][]byte {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_blobs(
		_pointer, _uniffiStatus)
	}))
}


func (_self *TransactionManifest)Compile() ([]byte, error) {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_compile(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *TransactionManifest)ExecutionSummary(networkId uint8, encodedReceipt []byte) (ExecutionSummary, error) {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_execution_summary(
		_pointer,FfiConverterUint8INSTANCE.Lower(networkId), FfiConverterBytesINSTANCE.Lower(encodedReceipt), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue ExecutionSummary
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeExecutionSummaryINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *TransactionManifest)ExtractAddresses() map[EntityType][]*Address {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterMapTypeEntityTypeSequenceAddressINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_extract_addresses(
		_pointer, _uniffiStatus)
	}))
}


func (_self *TransactionManifest)Instructions() *Instructions {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterInstructionsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_instructions(
		_pointer, _uniffiStatus)
	}))
}


func (_self *TransactionManifest)Modify(modifications TransactionManifestModifications) (*TransactionManifest, error) {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_modify(
		_pointer,FfiConverterTypeTransactionManifestModificationsINSTANCE.Lower(modifications), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *TransactionManifest
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTransactionManifestINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *TransactionManifest)StaticallyValidate() error {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_statically_validate(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


func (_self *TransactionManifest)Summary(networkId uint8) ManifestSummary {
	_pointer := _self.ffiObject.incrementPointer("*TransactionManifest")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeManifestSummaryINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_transactionmanifest_summary(
		_pointer,FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	}))
}



func (object *TransactionManifest)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTransactionManifest struct {}

var FfiConverterTransactionManifestINSTANCE = FfiConverterTransactionManifest{}

func (c FfiConverterTransactionManifest) Lift(pointer unsafe.Pointer) *TransactionManifest {
	result := &TransactionManifest {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_transactionmanifest(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*TransactionManifest).Destroy)
	return result
}

func (c FfiConverterTransactionManifest) Read(reader io.Reader) *TransactionManifest {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTransactionManifest) Lower(value *TransactionManifest) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TransactionManifest")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterTransactionManifest) Write(writer io.Writer, value *TransactionManifest) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTransactionManifest struct {}

func (_ FfiDestroyerTransactionManifest) Destroy(value *TransactionManifest) {
	value.Destroy()
}


type ValidationConfig struct {
	ffiObject FfiObject
}
func NewValidationConfig(networkId uint8, maxNotarizedPayloadSize uint64, minTipPercentage uint16, maxTipPercentage uint16, maxEpochRange uint64, messageValidation *MessageValidationConfig) *ValidationConfig {
	return FfiConverterValidationConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_validationconfig_new(FfiConverterUint8INSTANCE.Lower(networkId), FfiConverterUint64INSTANCE.Lower(maxNotarizedPayloadSize), FfiConverterUint16INSTANCE.Lower(minTipPercentage), FfiConverterUint16INSTANCE.Lower(maxTipPercentage), FfiConverterUint64INSTANCE.Lower(maxEpochRange), FfiConverterMessageValidationConfigINSTANCE.Lower(messageValidation), _uniffiStatus)
	}))
}


func ValidationConfigDefault(networkId uint8) *ValidationConfig {
	return FfiConverterValidationConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_constructor_validationconfig_default(FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	}))
}



func (_self *ValidationConfig)MaxEpochRange() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*ValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_validationconfig_max_epoch_range(
		_pointer, _uniffiStatus)
	}))
}


func (_self *ValidationConfig)MaxNotarizedPayloadSize() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*ValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_validationconfig_max_notarized_payload_size(
		_pointer, _uniffiStatus)
	}))
}


func (_self *ValidationConfig)MaxTipPercentage() uint16 {
	_pointer := _self.ffiObject.incrementPointer("*ValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint16INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_validationconfig_max_tip_percentage(
		_pointer, _uniffiStatus)
	}))
}


func (_self *ValidationConfig)MessageValidation() *MessageValidationConfig {
	_pointer := _self.ffiObject.incrementPointer("*ValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterMessageValidationConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_validationconfig_message_validation(
		_pointer, _uniffiStatus)
	}))
}


func (_self *ValidationConfig)MinTipPercentage() uint16 {
	_pointer := _self.ffiObject.incrementPointer("*ValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint16INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_validationconfig_min_tip_percentage(
		_pointer, _uniffiStatus)
	}))
}


func (_self *ValidationConfig)NetworkId() uint8 {
	_pointer := _self.ffiObject.incrementPointer("*ValidationConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint8INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint8_t {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_method_validationconfig_network_id(
		_pointer, _uniffiStatus)
	}))
}



func (object *ValidationConfig)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterValidationConfig struct {}

var FfiConverterValidationConfigINSTANCE = FfiConverterValidationConfig{}

func (c FfiConverterValidationConfig) Lift(pointer unsafe.Pointer) *ValidationConfig {
	result := &ValidationConfig {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_radix_engine_toolkit_uniffi_fn_free_validationconfig(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*ValidationConfig).Destroy)
	return result
}

func (c FfiConverterValidationConfig) Read(reader io.Reader) *ValidationConfig {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterValidationConfig) Lower(value *ValidationConfig) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*ValidationConfig")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterValidationConfig) Write(writer io.Writer, value *ValidationConfig) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerValidationConfig struct {}

func (_ FfiDestroyerValidationConfig) Destroy(value *ValidationConfig) {
	value.Destroy()
}


type AccountAddAuthorizedDepositorEvent struct {
	AuthorizedDepositorBadge ResourceOrNonFungible
}

func (r *AccountAddAuthorizedDepositorEvent) Destroy() {
		FfiDestroyerTypeResourceOrNonFungible{}.Destroy(r.AuthorizedDepositorBadge);
}

type FfiConverterTypeAccountAddAuthorizedDepositorEvent struct {}

var FfiConverterTypeAccountAddAuthorizedDepositorEventINSTANCE = FfiConverterTypeAccountAddAuthorizedDepositorEvent{}

func (c FfiConverterTypeAccountAddAuthorizedDepositorEvent) Lift(rb RustBufferI) AccountAddAuthorizedDepositorEvent {
	return LiftFromRustBuffer[AccountAddAuthorizedDepositorEvent](c, rb)
}

func (c FfiConverterTypeAccountAddAuthorizedDepositorEvent) Read(reader io.Reader) AccountAddAuthorizedDepositorEvent {
	return AccountAddAuthorizedDepositorEvent {
			FfiConverterTypeResourceOrNonFungibleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAccountAddAuthorizedDepositorEvent) Lower(value AccountAddAuthorizedDepositorEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountAddAuthorizedDepositorEvent](c, value)
}

func (c FfiConverterTypeAccountAddAuthorizedDepositorEvent) Write(writer io.Writer, value AccountAddAuthorizedDepositorEvent) {
		FfiConverterTypeResourceOrNonFungibleINSTANCE.Write(writer, value.AuthorizedDepositorBadge);
}

type FfiDestroyerTypeAccountAddAuthorizedDepositorEvent struct {}

func (_ FfiDestroyerTypeAccountAddAuthorizedDepositorEvent) Destroy(value AccountAddAuthorizedDepositorEvent) {
	value.Destroy()
}


type AccountRemoveAuthorizedDepositorEvent struct {
	AuthorizedDepositorBadge ResourceOrNonFungible
}

func (r *AccountRemoveAuthorizedDepositorEvent) Destroy() {
		FfiDestroyerTypeResourceOrNonFungible{}.Destroy(r.AuthorizedDepositorBadge);
}

type FfiConverterTypeAccountRemoveAuthorizedDepositorEvent struct {}

var FfiConverterTypeAccountRemoveAuthorizedDepositorEventINSTANCE = FfiConverterTypeAccountRemoveAuthorizedDepositorEvent{}

func (c FfiConverterTypeAccountRemoveAuthorizedDepositorEvent) Lift(rb RustBufferI) AccountRemoveAuthorizedDepositorEvent {
	return LiftFromRustBuffer[AccountRemoveAuthorizedDepositorEvent](c, rb)
}

func (c FfiConverterTypeAccountRemoveAuthorizedDepositorEvent) Read(reader io.Reader) AccountRemoveAuthorizedDepositorEvent {
	return AccountRemoveAuthorizedDepositorEvent {
			FfiConverterTypeResourceOrNonFungibleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAccountRemoveAuthorizedDepositorEvent) Lower(value AccountRemoveAuthorizedDepositorEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountRemoveAuthorizedDepositorEvent](c, value)
}

func (c FfiConverterTypeAccountRemoveAuthorizedDepositorEvent) Write(writer io.Writer, value AccountRemoveAuthorizedDepositorEvent) {
		FfiConverterTypeResourceOrNonFungibleINSTANCE.Write(writer, value.AuthorizedDepositorBadge);
}

type FfiDestroyerTypeAccountRemoveAuthorizedDepositorEvent struct {}

func (_ FfiDestroyerTypeAccountRemoveAuthorizedDepositorEvent) Destroy(value AccountRemoveAuthorizedDepositorEvent) {
	value.Destroy()
}


type AccountRemoveResourcePreferenceEvent struct {
	ResourceAddress *Address
}

func (r *AccountRemoveResourcePreferenceEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
}

type FfiConverterTypeAccountRemoveResourcePreferenceEvent struct {}

var FfiConverterTypeAccountRemoveResourcePreferenceEventINSTANCE = FfiConverterTypeAccountRemoveResourcePreferenceEvent{}

func (c FfiConverterTypeAccountRemoveResourcePreferenceEvent) Lift(rb RustBufferI) AccountRemoveResourcePreferenceEvent {
	return LiftFromRustBuffer[AccountRemoveResourcePreferenceEvent](c, rb)
}

func (c FfiConverterTypeAccountRemoveResourcePreferenceEvent) Read(reader io.Reader) AccountRemoveResourcePreferenceEvent {
	return AccountRemoveResourcePreferenceEvent {
			FfiConverterAddressINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAccountRemoveResourcePreferenceEvent) Lower(value AccountRemoveResourcePreferenceEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountRemoveResourcePreferenceEvent](c, value)
}

func (c FfiConverterTypeAccountRemoveResourcePreferenceEvent) Write(writer io.Writer, value AccountRemoveResourcePreferenceEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
}

type FfiDestroyerTypeAccountRemoveResourcePreferenceEvent struct {}

func (_ FfiDestroyerTypeAccountRemoveResourcePreferenceEvent) Destroy(value AccountRemoveResourcePreferenceEvent) {
	value.Destroy()
}


type AccountSetDefaultDepositRuleEvent struct {
	DefaultDepositRule AccountDefaultDepositRule
}

func (r *AccountSetDefaultDepositRuleEvent) Destroy() {
		FfiDestroyerTypeAccountDefaultDepositRule{}.Destroy(r.DefaultDepositRule);
}

type FfiConverterTypeAccountSetDefaultDepositRuleEvent struct {}

var FfiConverterTypeAccountSetDefaultDepositRuleEventINSTANCE = FfiConverterTypeAccountSetDefaultDepositRuleEvent{}

func (c FfiConverterTypeAccountSetDefaultDepositRuleEvent) Lift(rb RustBufferI) AccountSetDefaultDepositRuleEvent {
	return LiftFromRustBuffer[AccountSetDefaultDepositRuleEvent](c, rb)
}

func (c FfiConverterTypeAccountSetDefaultDepositRuleEvent) Read(reader io.Reader) AccountSetDefaultDepositRuleEvent {
	return AccountSetDefaultDepositRuleEvent {
			FfiConverterTypeAccountDefaultDepositRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAccountSetDefaultDepositRuleEvent) Lower(value AccountSetDefaultDepositRuleEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountSetDefaultDepositRuleEvent](c, value)
}

func (c FfiConverterTypeAccountSetDefaultDepositRuleEvent) Write(writer io.Writer, value AccountSetDefaultDepositRuleEvent) {
		FfiConverterTypeAccountDefaultDepositRuleINSTANCE.Write(writer, value.DefaultDepositRule);
}

type FfiDestroyerTypeAccountSetDefaultDepositRuleEvent struct {}

func (_ FfiDestroyerTypeAccountSetDefaultDepositRuleEvent) Destroy(value AccountSetDefaultDepositRuleEvent) {
	value.Destroy()
}


type AccountSetResourcePreferenceEvent struct {
	ResourceAddress *Address
	Preference ResourcePreference
}

func (r *AccountSetResourcePreferenceEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerTypeResourcePreference{}.Destroy(r.Preference);
}

type FfiConverterTypeAccountSetResourcePreferenceEvent struct {}

var FfiConverterTypeAccountSetResourcePreferenceEventINSTANCE = FfiConverterTypeAccountSetResourcePreferenceEvent{}

func (c FfiConverterTypeAccountSetResourcePreferenceEvent) Lift(rb RustBufferI) AccountSetResourcePreferenceEvent {
	return LiftFromRustBuffer[AccountSetResourcePreferenceEvent](c, rb)
}

func (c FfiConverterTypeAccountSetResourcePreferenceEvent) Read(reader io.Reader) AccountSetResourcePreferenceEvent {
	return AccountSetResourcePreferenceEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterTypeResourcePreferenceINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAccountSetResourcePreferenceEvent) Lower(value AccountSetResourcePreferenceEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountSetResourcePreferenceEvent](c, value)
}

func (c FfiConverterTypeAccountSetResourcePreferenceEvent) Write(writer io.Writer, value AccountSetResourcePreferenceEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterTypeResourcePreferenceINSTANCE.Write(writer, value.Preference);
}

type FfiDestroyerTypeAccountSetResourcePreferenceEvent struct {}

func (_ FfiDestroyerTypeAccountSetResourcePreferenceEvent) Destroy(value AccountSetResourcePreferenceEvent) {
	value.Destroy()
}


type BadgeWithdrawEvent struct {
	Proposer Proposer
}

func (r *BadgeWithdrawEvent) Destroy() {
		FfiDestroyerTypeProposer{}.Destroy(r.Proposer);
}

type FfiConverterTypeBadgeWithdrawEvent struct {}

var FfiConverterTypeBadgeWithdrawEventINSTANCE = FfiConverterTypeBadgeWithdrawEvent{}

func (c FfiConverterTypeBadgeWithdrawEvent) Lift(rb RustBufferI) BadgeWithdrawEvent {
	return LiftFromRustBuffer[BadgeWithdrawEvent](c, rb)
}

func (c FfiConverterTypeBadgeWithdrawEvent) Read(reader io.Reader) BadgeWithdrawEvent {
	return BadgeWithdrawEvent {
			FfiConverterTypeProposerINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBadgeWithdrawEvent) Lower(value BadgeWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[BadgeWithdrawEvent](c, value)
}

func (c FfiConverterTypeBadgeWithdrawEvent) Write(writer io.Writer, value BadgeWithdrawEvent) {
		FfiConverterTypeProposerINSTANCE.Write(writer, value.Proposer);
}

type FfiDestroyerTypeBadgeWithdrawEvent struct {}

func (_ FfiDestroyerTypeBadgeWithdrawEvent) Destroy(value BadgeWithdrawEvent) {
	value.Destroy()
}


type BuildInformation struct {
	Version string
	ScryptoDependency DependencyInformation
}

func (r *BuildInformation) Destroy() {
		FfiDestroyerString{}.Destroy(r.Version);
		FfiDestroyerTypeDependencyInformation{}.Destroy(r.ScryptoDependency);
}

type FfiConverterTypeBuildInformation struct {}

var FfiConverterTypeBuildInformationINSTANCE = FfiConverterTypeBuildInformation{}

func (c FfiConverterTypeBuildInformation) Lift(rb RustBufferI) BuildInformation {
	return LiftFromRustBuffer[BuildInformation](c, rb)
}

func (c FfiConverterTypeBuildInformation) Read(reader io.Reader) BuildInformation {
	return BuildInformation {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypeDependencyInformationINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBuildInformation) Lower(value BuildInformation) RustBuffer {
	return LowerIntoRustBuffer[BuildInformation](c, value)
}

func (c FfiConverterTypeBuildInformation) Write(writer io.Writer, value BuildInformation) {
		FfiConverterStringINSTANCE.Write(writer, value.Version);
		FfiConverterTypeDependencyInformationINSTANCE.Write(writer, value.ScryptoDependency);
}

type FfiDestroyerTypeBuildInformation struct {}

func (_ FfiDestroyerTypeBuildInformation) Destroy(value BuildInformation) {
	value.Destroy()
}


type BurnFungibleResourceEvent struct {
	Amount *Decimal
}

func (r *BurnFungibleResourceEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeBurnFungibleResourceEvent struct {}

var FfiConverterTypeBurnFungibleResourceEventINSTANCE = FfiConverterTypeBurnFungibleResourceEvent{}

func (c FfiConverterTypeBurnFungibleResourceEvent) Lift(rb RustBufferI) BurnFungibleResourceEvent {
	return LiftFromRustBuffer[BurnFungibleResourceEvent](c, rb)
}

func (c FfiConverterTypeBurnFungibleResourceEvent) Read(reader io.Reader) BurnFungibleResourceEvent {
	return BurnFungibleResourceEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBurnFungibleResourceEvent) Lower(value BurnFungibleResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[BurnFungibleResourceEvent](c, value)
}

func (c FfiConverterTypeBurnFungibleResourceEvent) Write(writer io.Writer, value BurnFungibleResourceEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeBurnFungibleResourceEvent struct {}

func (_ FfiDestroyerTypeBurnFungibleResourceEvent) Destroy(value BurnFungibleResourceEvent) {
	value.Destroy()
}


type BurnNonFungibleResourceEvent struct {
	Ids []NonFungibleLocalId
}

func (r *BurnNonFungibleResourceEvent) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.Ids);
}

type FfiConverterTypeBurnNonFungibleResourceEvent struct {}

var FfiConverterTypeBurnNonFungibleResourceEventINSTANCE = FfiConverterTypeBurnNonFungibleResourceEvent{}

func (c FfiConverterTypeBurnNonFungibleResourceEvent) Lift(rb RustBufferI) BurnNonFungibleResourceEvent {
	return LiftFromRustBuffer[BurnNonFungibleResourceEvent](c, rb)
}

func (c FfiConverterTypeBurnNonFungibleResourceEvent) Read(reader io.Reader) BurnNonFungibleResourceEvent {
	return BurnNonFungibleResourceEvent {
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBurnNonFungibleResourceEvent) Lower(value BurnNonFungibleResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[BurnNonFungibleResourceEvent](c, value)
}

func (c FfiConverterTypeBurnNonFungibleResourceEvent) Write(writer io.Writer, value BurnNonFungibleResourceEvent) {
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.Ids);
}

type FfiDestroyerTypeBurnNonFungibleResourceEvent struct {}

func (_ FfiDestroyerTypeBurnNonFungibleResourceEvent) Destroy(value BurnNonFungibleResourceEvent) {
	value.Destroy()
}


type CancelBadgeWithdrawAttemptEvent struct {
	Proposer Proposer
}

func (r *CancelBadgeWithdrawAttemptEvent) Destroy() {
		FfiDestroyerTypeProposer{}.Destroy(r.Proposer);
}

type FfiConverterTypeCancelBadgeWithdrawAttemptEvent struct {}

var FfiConverterTypeCancelBadgeWithdrawAttemptEventINSTANCE = FfiConverterTypeCancelBadgeWithdrawAttemptEvent{}

func (c FfiConverterTypeCancelBadgeWithdrawAttemptEvent) Lift(rb RustBufferI) CancelBadgeWithdrawAttemptEvent {
	return LiftFromRustBuffer[CancelBadgeWithdrawAttemptEvent](c, rb)
}

func (c FfiConverterTypeCancelBadgeWithdrawAttemptEvent) Read(reader io.Reader) CancelBadgeWithdrawAttemptEvent {
	return CancelBadgeWithdrawAttemptEvent {
			FfiConverterTypeProposerINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeCancelBadgeWithdrawAttemptEvent) Lower(value CancelBadgeWithdrawAttemptEvent) RustBuffer {
	return LowerIntoRustBuffer[CancelBadgeWithdrawAttemptEvent](c, value)
}

func (c FfiConverterTypeCancelBadgeWithdrawAttemptEvent) Write(writer io.Writer, value CancelBadgeWithdrawAttemptEvent) {
		FfiConverterTypeProposerINSTANCE.Write(writer, value.Proposer);
}

type FfiDestroyerTypeCancelBadgeWithdrawAttemptEvent struct {}

func (_ FfiDestroyerTypeCancelBadgeWithdrawAttemptEvent) Destroy(value CancelBadgeWithdrawAttemptEvent) {
	value.Destroy()
}


type CancelRecoveryProposalEvent struct {
	Proposer Proposer
}

func (r *CancelRecoveryProposalEvent) Destroy() {
		FfiDestroyerTypeProposer{}.Destroy(r.Proposer);
}

type FfiConverterTypeCancelRecoveryProposalEvent struct {}

var FfiConverterTypeCancelRecoveryProposalEventINSTANCE = FfiConverterTypeCancelRecoveryProposalEvent{}

func (c FfiConverterTypeCancelRecoveryProposalEvent) Lift(rb RustBufferI) CancelRecoveryProposalEvent {
	return LiftFromRustBuffer[CancelRecoveryProposalEvent](c, rb)
}

func (c FfiConverterTypeCancelRecoveryProposalEvent) Read(reader io.Reader) CancelRecoveryProposalEvent {
	return CancelRecoveryProposalEvent {
			FfiConverterTypeProposerINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeCancelRecoveryProposalEvent) Lower(value CancelRecoveryProposalEvent) RustBuffer {
	return LowerIntoRustBuffer[CancelRecoveryProposalEvent](c, value)
}

func (c FfiConverterTypeCancelRecoveryProposalEvent) Write(writer io.Writer, value CancelRecoveryProposalEvent) {
		FfiConverterTypeProposerINSTANCE.Write(writer, value.Proposer);
}

type FfiDestroyerTypeCancelRecoveryProposalEvent struct {}

func (_ FfiDestroyerTypeCancelRecoveryProposalEvent) Destroy(value CancelRecoveryProposalEvent) {
	value.Destroy()
}


type ClaimEvent struct {
	Claimant *Address
	ResourceAddress *Address
	Resources ResourceSpecifier
}

func (r *ClaimEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.Claimant);
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerTypeResourceSpecifier{}.Destroy(r.Resources);
}

type FfiConverterTypeClaimEvent struct {}

var FfiConverterTypeClaimEventINSTANCE = FfiConverterTypeClaimEvent{}

func (c FfiConverterTypeClaimEvent) Lift(rb RustBufferI) ClaimEvent {
	return LiftFromRustBuffer[ClaimEvent](c, rb)
}

func (c FfiConverterTypeClaimEvent) Read(reader io.Reader) ClaimEvent {
	return ClaimEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterTypeResourceSpecifierINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeClaimEvent) Lower(value ClaimEvent) RustBuffer {
	return LowerIntoRustBuffer[ClaimEvent](c, value)
}

func (c FfiConverterTypeClaimEvent) Write(writer io.Writer, value ClaimEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.Claimant);
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterTypeResourceSpecifierINSTANCE.Write(writer, value.Resources);
}

type FfiDestroyerTypeClaimEvent struct {}

func (_ FfiDestroyerTypeClaimEvent) Destroy(value ClaimEvent) {
	value.Destroy()
}


type ClaimXrdEvent struct {
	ClaimedXrd *Decimal
}

func (r *ClaimXrdEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.ClaimedXrd);
}

type FfiConverterTypeClaimXrdEvent struct {}

var FfiConverterTypeClaimXrdEventINSTANCE = FfiConverterTypeClaimXrdEvent{}

func (c FfiConverterTypeClaimXrdEvent) Lift(rb RustBufferI) ClaimXrdEvent {
	return LiftFromRustBuffer[ClaimXrdEvent](c, rb)
}

func (c FfiConverterTypeClaimXrdEvent) Read(reader io.Reader) ClaimXrdEvent {
	return ClaimXrdEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeClaimXrdEvent) Lower(value ClaimXrdEvent) RustBuffer {
	return LowerIntoRustBuffer[ClaimXrdEvent](c, value)
}

func (c FfiConverterTypeClaimXrdEvent) Write(writer io.Writer, value ClaimXrdEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.ClaimedXrd);
}

type FfiDestroyerTypeClaimXrdEvent struct {}

func (_ FfiDestroyerTypeClaimXrdEvent) Destroy(value ClaimXrdEvent) {
	value.Destroy()
}


type ComponentAddresses struct {
	ConsensusManager *Address
	GenesisHelper *Address
	Faucet *Address
}

func (r *ComponentAddresses) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ConsensusManager);
		FfiDestroyerAddress{}.Destroy(r.GenesisHelper);
		FfiDestroyerAddress{}.Destroy(r.Faucet);
}

type FfiConverterTypeComponentAddresses struct {}

var FfiConverterTypeComponentAddressesINSTANCE = FfiConverterTypeComponentAddresses{}

func (c FfiConverterTypeComponentAddresses) Lift(rb RustBufferI) ComponentAddresses {
	return LiftFromRustBuffer[ComponentAddresses](c, rb)
}

func (c FfiConverterTypeComponentAddresses) Read(reader io.Reader) ComponentAddresses {
	return ComponentAddresses {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeComponentAddresses) Lower(value ComponentAddresses) RustBuffer {
	return LowerIntoRustBuffer[ComponentAddresses](c, value)
}

func (c FfiConverterTypeComponentAddresses) Write(writer io.Writer, value ComponentAddresses) {
		FfiConverterAddressINSTANCE.Write(writer, value.ConsensusManager);
		FfiConverterAddressINSTANCE.Write(writer, value.GenesisHelper);
		FfiConverterAddressINSTANCE.Write(writer, value.Faucet);
}

type FfiDestroyerTypeComponentAddresses struct {}

func (_ FfiDestroyerTypeComponentAddresses) Destroy(value ComponentAddresses) {
	value.Destroy()
}


type Ed25519PublicKey struct {
	Value []byte
}

func (r *Ed25519PublicKey) Destroy() {
		FfiDestroyerBytes{}.Destroy(r.Value);
}

type FfiConverterTypeEd25519PublicKey struct {}

var FfiConverterTypeEd25519PublicKeyINSTANCE = FfiConverterTypeEd25519PublicKey{}

func (c FfiConverterTypeEd25519PublicKey) Lift(rb RustBufferI) Ed25519PublicKey {
	return LiftFromRustBuffer[Ed25519PublicKey](c, rb)
}

func (c FfiConverterTypeEd25519PublicKey) Read(reader io.Reader) Ed25519PublicKey {
	return Ed25519PublicKey {
			FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeEd25519PublicKey) Lower(value Ed25519PublicKey) RustBuffer {
	return LowerIntoRustBuffer[Ed25519PublicKey](c, value)
}

func (c FfiConverterTypeEd25519PublicKey) Write(writer io.Writer, value Ed25519PublicKey) {
		FfiConverterBytesINSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeEd25519PublicKey struct {}

func (_ FfiDestroyerTypeEd25519PublicKey) Destroy(value Ed25519PublicKey) {
	value.Destroy()
}


type EncryptedMessage struct {
	Encrypted []byte
	DecryptorsByCurve map[CurveType]DecryptorsByCurve
}

func (r *EncryptedMessage) Destroy() {
		FfiDestroyerBytes{}.Destroy(r.Encrypted);
		FfiDestroyerMapTypeCurveTypeTypeDecryptorsByCurve{}.Destroy(r.DecryptorsByCurve);
}

type FfiConverterTypeEncryptedMessage struct {}

var FfiConverterTypeEncryptedMessageINSTANCE = FfiConverterTypeEncryptedMessage{}

func (c FfiConverterTypeEncryptedMessage) Lift(rb RustBufferI) EncryptedMessage {
	return LiftFromRustBuffer[EncryptedMessage](c, rb)
}

func (c FfiConverterTypeEncryptedMessage) Read(reader io.Reader) EncryptedMessage {
	return EncryptedMessage {
			FfiConverterBytesINSTANCE.Read(reader),
			FfiConverterMapTypeCurveTypeTypeDecryptorsByCurveINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeEncryptedMessage) Lower(value EncryptedMessage) RustBuffer {
	return LowerIntoRustBuffer[EncryptedMessage](c, value)
}

func (c FfiConverterTypeEncryptedMessage) Write(writer io.Writer, value EncryptedMessage) {
		FfiConverterBytesINSTANCE.Write(writer, value.Encrypted);
		FfiConverterMapTypeCurveTypeTypeDecryptorsByCurveINSTANCE.Write(writer, value.DecryptorsByCurve);
}

type FfiDestroyerTypeEncryptedMessage struct {}

func (_ FfiDestroyerTypeEncryptedMessage) Destroy(value EncryptedMessage) {
	value.Destroy()
}


type EpochChangeEvent struct {
	Epoch uint64
	ValidatorSet map[string]ValidatorInfo
}

func (r *EpochChangeEvent) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.Epoch);
		FfiDestroyerMapStringTypeValidatorInfo{}.Destroy(r.ValidatorSet);
}

type FfiConverterTypeEpochChangeEvent struct {}

var FfiConverterTypeEpochChangeEventINSTANCE = FfiConverterTypeEpochChangeEvent{}

func (c FfiConverterTypeEpochChangeEvent) Lift(rb RustBufferI) EpochChangeEvent {
	return LiftFromRustBuffer[EpochChangeEvent](c, rb)
}

func (c FfiConverterTypeEpochChangeEvent) Read(reader io.Reader) EpochChangeEvent {
	return EpochChangeEvent {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterMapStringTypeValidatorInfoINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeEpochChangeEvent) Lower(value EpochChangeEvent) RustBuffer {
	return LowerIntoRustBuffer[EpochChangeEvent](c, value)
}

func (c FfiConverterTypeEpochChangeEvent) Write(writer io.Writer, value EpochChangeEvent) {
		FfiConverterUint64INSTANCE.Write(writer, value.Epoch);
		FfiConverterMapStringTypeValidatorInfoINSTANCE.Write(writer, value.ValidatorSet);
}

type FfiDestroyerTypeEpochChangeEvent struct {}

func (_ FfiDestroyerTypeEpochChangeEvent) Destroy(value EpochChangeEvent) {
	value.Destroy()
}


type EventTypeIdentifier struct {
	Emitter Emitter
	EventName string
}

func (r *EventTypeIdentifier) Destroy() {
		FfiDestroyerTypeEmitter{}.Destroy(r.Emitter);
		FfiDestroyerString{}.Destroy(r.EventName);
}

type FfiConverterTypeEventTypeIdentifier struct {}

var FfiConverterTypeEventTypeIdentifierINSTANCE = FfiConverterTypeEventTypeIdentifier{}

func (c FfiConverterTypeEventTypeIdentifier) Lift(rb RustBufferI) EventTypeIdentifier {
	return LiftFromRustBuffer[EventTypeIdentifier](c, rb)
}

func (c FfiConverterTypeEventTypeIdentifier) Read(reader io.Reader) EventTypeIdentifier {
	return EventTypeIdentifier {
			FfiConverterTypeEmitterINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeEventTypeIdentifier) Lower(value EventTypeIdentifier) RustBuffer {
	return LowerIntoRustBuffer[EventTypeIdentifier](c, value)
}

func (c FfiConverterTypeEventTypeIdentifier) Write(writer io.Writer, value EventTypeIdentifier) {
		FfiConverterTypeEmitterINSTANCE.Write(writer, value.Emitter);
		FfiConverterStringINSTANCE.Write(writer, value.EventName);
}

type FfiDestroyerTypeEventTypeIdentifier struct {}

func (_ FfiDestroyerTypeEventTypeIdentifier) Destroy(value EventTypeIdentifier) {
	value.Destroy()
}


type ExecutionSummary struct {
	AccountWithdraws map[string][]ResourceIndicator
	AccountDeposits map[string][]ResourceIndicator
	PresentedProofs map[string][]ResourceSpecifier
	NewEntities NewEntities
	EncounteredEntities []*Address
	AccountsRequiringAuth []*Address
	IdentitiesRequiringAuth []*Address
	ReservedInstructions []ReservedInstruction
	FeeLocks FeeLocks
	FeeSummary FeeSummary
	DetailedClassification []DetailedManifestClass
	NewlyCreatedNonFungibles []*NonFungibleGlobalId
}

func (r *ExecutionSummary) Destroy() {
		FfiDestroyerMapStringSequenceTypeResourceIndicator{}.Destroy(r.AccountWithdraws);
		FfiDestroyerMapStringSequenceTypeResourceIndicator{}.Destroy(r.AccountDeposits);
		FfiDestroyerMapStringSequenceTypeResourceSpecifier{}.Destroy(r.PresentedProofs);
		FfiDestroyerTypeNewEntities{}.Destroy(r.NewEntities);
		FfiDestroyerSequenceAddress{}.Destroy(r.EncounteredEntities);
		FfiDestroyerSequenceAddress{}.Destroy(r.AccountsRequiringAuth);
		FfiDestroyerSequenceAddress{}.Destroy(r.IdentitiesRequiringAuth);
		FfiDestroyerSequenceTypeReservedInstruction{}.Destroy(r.ReservedInstructions);
		FfiDestroyerTypeFeeLocks{}.Destroy(r.FeeLocks);
		FfiDestroyerTypeFeeSummary{}.Destroy(r.FeeSummary);
		FfiDestroyerSequenceTypeDetailedManifestClass{}.Destroy(r.DetailedClassification);
		FfiDestroyerSequenceNonFungibleGlobalId{}.Destroy(r.NewlyCreatedNonFungibles);
}

type FfiConverterTypeExecutionSummary struct {}

var FfiConverterTypeExecutionSummaryINSTANCE = FfiConverterTypeExecutionSummary{}

func (c FfiConverterTypeExecutionSummary) Lift(rb RustBufferI) ExecutionSummary {
	return LiftFromRustBuffer[ExecutionSummary](c, rb)
}

func (c FfiConverterTypeExecutionSummary) Read(reader io.Reader) ExecutionSummary {
	return ExecutionSummary {
			FfiConverterMapStringSequenceTypeResourceIndicatorINSTANCE.Read(reader),
			FfiConverterMapStringSequenceTypeResourceIndicatorINSTANCE.Read(reader),
			FfiConverterMapStringSequenceTypeResourceSpecifierINSTANCE.Read(reader),
			FfiConverterTypeNewEntitiesINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceTypeReservedInstructionINSTANCE.Read(reader),
			FfiConverterTypeFeeLocksINSTANCE.Read(reader),
			FfiConverterTypeFeeSummaryINSTANCE.Read(reader),
			FfiConverterSequenceTypeDetailedManifestClassINSTANCE.Read(reader),
			FfiConverterSequenceNonFungibleGlobalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeExecutionSummary) Lower(value ExecutionSummary) RustBuffer {
	return LowerIntoRustBuffer[ExecutionSummary](c, value)
}

func (c FfiConverterTypeExecutionSummary) Write(writer io.Writer, value ExecutionSummary) {
		FfiConverterMapStringSequenceTypeResourceIndicatorINSTANCE.Write(writer, value.AccountWithdraws);
		FfiConverterMapStringSequenceTypeResourceIndicatorINSTANCE.Write(writer, value.AccountDeposits);
		FfiConverterMapStringSequenceTypeResourceSpecifierINSTANCE.Write(writer, value.PresentedProofs);
		FfiConverterTypeNewEntitiesINSTANCE.Write(writer, value.NewEntities);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.EncounteredEntities);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.AccountsRequiringAuth);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.IdentitiesRequiringAuth);
		FfiConverterSequenceTypeReservedInstructionINSTANCE.Write(writer, value.ReservedInstructions);
		FfiConverterTypeFeeLocksINSTANCE.Write(writer, value.FeeLocks);
		FfiConverterTypeFeeSummaryINSTANCE.Write(writer, value.FeeSummary);
		FfiConverterSequenceTypeDetailedManifestClassINSTANCE.Write(writer, value.DetailedClassification);
		FfiConverterSequenceNonFungibleGlobalIdINSTANCE.Write(writer, value.NewlyCreatedNonFungibles);
}

type FfiDestroyerTypeExecutionSummary struct {}

func (_ FfiDestroyerTypeExecutionSummary) Destroy(value ExecutionSummary) {
	value.Destroy()
}


type FeeLocks struct {
	Lock *Decimal
	ContingentLock *Decimal
}

func (r *FeeLocks) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Lock);
		FfiDestroyerDecimal{}.Destroy(r.ContingentLock);
}

type FfiConverterTypeFeeLocks struct {}

var FfiConverterTypeFeeLocksINSTANCE = FfiConverterTypeFeeLocks{}

func (c FfiConverterTypeFeeLocks) Lift(rb RustBufferI) FeeLocks {
	return LiftFromRustBuffer[FeeLocks](c, rb)
}

func (c FfiConverterTypeFeeLocks) Read(reader io.Reader) FeeLocks {
	return FeeLocks {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeeLocks) Lower(value FeeLocks) RustBuffer {
	return LowerIntoRustBuffer[FeeLocks](c, value)
}

func (c FfiConverterTypeFeeLocks) Write(writer io.Writer, value FeeLocks) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Lock);
		FfiConverterDecimalINSTANCE.Write(writer, value.ContingentLock);
}

type FfiDestroyerTypeFeeLocks struct {}

func (_ FfiDestroyerTypeFeeLocks) Destroy(value FeeLocks) {
	value.Destroy()
}


type FeeSummary struct {
	ExecutionCost *Decimal
	FinalizationCost *Decimal
	StorageExpansionCost *Decimal
	RoyaltyCost *Decimal
}

func (r *FeeSummary) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.ExecutionCost);
		FfiDestroyerDecimal{}.Destroy(r.FinalizationCost);
		FfiDestroyerDecimal{}.Destroy(r.StorageExpansionCost);
		FfiDestroyerDecimal{}.Destroy(r.RoyaltyCost);
}

type FfiConverterTypeFeeSummary struct {}

var FfiConverterTypeFeeSummaryINSTANCE = FfiConverterTypeFeeSummary{}

func (c FfiConverterTypeFeeSummary) Lift(rb RustBufferI) FeeSummary {
	return LiftFromRustBuffer[FeeSummary](c, rb)
}

func (c FfiConverterTypeFeeSummary) Read(reader io.Reader) FeeSummary {
	return FeeSummary {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeeSummary) Lower(value FeeSummary) RustBuffer {
	return LowerIntoRustBuffer[FeeSummary](c, value)
}

func (c FfiConverterTypeFeeSummary) Write(writer io.Writer, value FeeSummary) {
		FfiConverterDecimalINSTANCE.Write(writer, value.ExecutionCost);
		FfiConverterDecimalINSTANCE.Write(writer, value.FinalizationCost);
		FfiConverterDecimalINSTANCE.Write(writer, value.StorageExpansionCost);
		FfiConverterDecimalINSTANCE.Write(writer, value.RoyaltyCost);
}

type FfiDestroyerTypeFeeSummary struct {}

func (_ FfiDestroyerTypeFeeSummary) Destroy(value FeeSummary) {
	value.Destroy()
}


type FungibleResourceRoles struct {
	MintRoles *ResourceManagerRole
	BurnRoles *ResourceManagerRole
	FreezeRoles *ResourceManagerRole
	RecallRoles *ResourceManagerRole
	WithdrawRoles *ResourceManagerRole
	DepositRoles *ResourceManagerRole
}

func (r *FungibleResourceRoles) Destroy() {
		FfiDestroyerOptionalTypeResourceManagerRole{}.Destroy(r.MintRoles);
		FfiDestroyerOptionalTypeResourceManagerRole{}.Destroy(r.BurnRoles);
		FfiDestroyerOptionalTypeResourceManagerRole{}.Destroy(r.FreezeRoles);
		FfiDestroyerOptionalTypeResourceManagerRole{}.Destroy(r.RecallRoles);
		FfiDestroyerOptionalTypeResourceManagerRole{}.Destroy(r.WithdrawRoles);
		FfiDestroyerOptionalTypeResourceManagerRole{}.Destroy(r.DepositRoles);
}

type FfiConverterTypeFungibleResourceRoles struct {}

var FfiConverterTypeFungibleResourceRolesINSTANCE = FfiConverterTypeFungibleResourceRoles{}

func (c FfiConverterTypeFungibleResourceRoles) Lift(rb RustBufferI) FungibleResourceRoles {
	return LiftFromRustBuffer[FungibleResourceRoles](c, rb)
}

func (c FfiConverterTypeFungibleResourceRoles) Read(reader io.Reader) FungibleResourceRoles {
	return FungibleResourceRoles {
			FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Read(reader),
			FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Read(reader),
			FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Read(reader),
			FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Read(reader),
			FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Read(reader),
			FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFungibleResourceRoles) Lower(value FungibleResourceRoles) RustBuffer {
	return LowerIntoRustBuffer[FungibleResourceRoles](c, value)
}

func (c FfiConverterTypeFungibleResourceRoles) Write(writer io.Writer, value FungibleResourceRoles) {
		FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Write(writer, value.MintRoles);
		FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Write(writer, value.BurnRoles);
		FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Write(writer, value.FreezeRoles);
		FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Write(writer, value.RecallRoles);
		FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Write(writer, value.WithdrawRoles);
		FfiConverterOptionalTypeResourceManagerRoleINSTANCE.Write(writer, value.DepositRoles);
}

type FfiDestroyerTypeFungibleResourceRoles struct {}

func (_ FfiDestroyerTypeFungibleResourceRoles) Destroy(value FungibleResourceRoles) {
	value.Destroy()
}


type FungibleVaultDepositEvent struct {
	Amount *Decimal
}

func (r *FungibleVaultDepositEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeFungibleVaultDepositEvent struct {}

var FfiConverterTypeFungibleVaultDepositEventINSTANCE = FfiConverterTypeFungibleVaultDepositEvent{}

func (c FfiConverterTypeFungibleVaultDepositEvent) Lift(rb RustBufferI) FungibleVaultDepositEvent {
	return LiftFromRustBuffer[FungibleVaultDepositEvent](c, rb)
}

func (c FfiConverterTypeFungibleVaultDepositEvent) Read(reader io.Reader) FungibleVaultDepositEvent {
	return FungibleVaultDepositEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFungibleVaultDepositEvent) Lower(value FungibleVaultDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[FungibleVaultDepositEvent](c, value)
}

func (c FfiConverterTypeFungibleVaultDepositEvent) Write(writer io.Writer, value FungibleVaultDepositEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeFungibleVaultDepositEvent struct {}

func (_ FfiDestroyerTypeFungibleVaultDepositEvent) Destroy(value FungibleVaultDepositEvent) {
	value.Destroy()
}


type FungibleVaultLockFeeEvent struct {
	Amount *Decimal
}

func (r *FungibleVaultLockFeeEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeFungibleVaultLockFeeEvent struct {}

var FfiConverterTypeFungibleVaultLockFeeEventINSTANCE = FfiConverterTypeFungibleVaultLockFeeEvent{}

func (c FfiConverterTypeFungibleVaultLockFeeEvent) Lift(rb RustBufferI) FungibleVaultLockFeeEvent {
	return LiftFromRustBuffer[FungibleVaultLockFeeEvent](c, rb)
}

func (c FfiConverterTypeFungibleVaultLockFeeEvent) Read(reader io.Reader) FungibleVaultLockFeeEvent {
	return FungibleVaultLockFeeEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFungibleVaultLockFeeEvent) Lower(value FungibleVaultLockFeeEvent) RustBuffer {
	return LowerIntoRustBuffer[FungibleVaultLockFeeEvent](c, value)
}

func (c FfiConverterTypeFungibleVaultLockFeeEvent) Write(writer io.Writer, value FungibleVaultLockFeeEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeFungibleVaultLockFeeEvent struct {}

func (_ FfiDestroyerTypeFungibleVaultLockFeeEvent) Destroy(value FungibleVaultLockFeeEvent) {
	value.Destroy()
}


type FungibleVaultPayFeeEvent struct {
	Amount *Decimal
}

func (r *FungibleVaultPayFeeEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeFungibleVaultPayFeeEvent struct {}

var FfiConverterTypeFungibleVaultPayFeeEventINSTANCE = FfiConverterTypeFungibleVaultPayFeeEvent{}

func (c FfiConverterTypeFungibleVaultPayFeeEvent) Lift(rb RustBufferI) FungibleVaultPayFeeEvent {
	return LiftFromRustBuffer[FungibleVaultPayFeeEvent](c, rb)
}

func (c FfiConverterTypeFungibleVaultPayFeeEvent) Read(reader io.Reader) FungibleVaultPayFeeEvent {
	return FungibleVaultPayFeeEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFungibleVaultPayFeeEvent) Lower(value FungibleVaultPayFeeEvent) RustBuffer {
	return LowerIntoRustBuffer[FungibleVaultPayFeeEvent](c, value)
}

func (c FfiConverterTypeFungibleVaultPayFeeEvent) Write(writer io.Writer, value FungibleVaultPayFeeEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeFungibleVaultPayFeeEvent struct {}

func (_ FfiDestroyerTypeFungibleVaultPayFeeEvent) Destroy(value FungibleVaultPayFeeEvent) {
	value.Destroy()
}


type FungibleVaultRecallEvent struct {
	Amount *Decimal
}

func (r *FungibleVaultRecallEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeFungibleVaultRecallEvent struct {}

var FfiConverterTypeFungibleVaultRecallEventINSTANCE = FfiConverterTypeFungibleVaultRecallEvent{}

func (c FfiConverterTypeFungibleVaultRecallEvent) Lift(rb RustBufferI) FungibleVaultRecallEvent {
	return LiftFromRustBuffer[FungibleVaultRecallEvent](c, rb)
}

func (c FfiConverterTypeFungibleVaultRecallEvent) Read(reader io.Reader) FungibleVaultRecallEvent {
	return FungibleVaultRecallEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFungibleVaultRecallEvent) Lower(value FungibleVaultRecallEvent) RustBuffer {
	return LowerIntoRustBuffer[FungibleVaultRecallEvent](c, value)
}

func (c FfiConverterTypeFungibleVaultRecallEvent) Write(writer io.Writer, value FungibleVaultRecallEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeFungibleVaultRecallEvent struct {}

func (_ FfiDestroyerTypeFungibleVaultRecallEvent) Destroy(value FungibleVaultRecallEvent) {
	value.Destroy()
}


type FungibleVaultWithdrawEvent struct {
	Amount *Decimal
}

func (r *FungibleVaultWithdrawEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeFungibleVaultWithdrawEvent struct {}

var FfiConverterTypeFungibleVaultWithdrawEventINSTANCE = FfiConverterTypeFungibleVaultWithdrawEvent{}

func (c FfiConverterTypeFungibleVaultWithdrawEvent) Lift(rb RustBufferI) FungibleVaultWithdrawEvent {
	return LiftFromRustBuffer[FungibleVaultWithdrawEvent](c, rb)
}

func (c FfiConverterTypeFungibleVaultWithdrawEvent) Read(reader io.Reader) FungibleVaultWithdrawEvent {
	return FungibleVaultWithdrawEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFungibleVaultWithdrawEvent) Lower(value FungibleVaultWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[FungibleVaultWithdrawEvent](c, value)
}

func (c FfiConverterTypeFungibleVaultWithdrawEvent) Write(writer io.Writer, value FungibleVaultWithdrawEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeFungibleVaultWithdrawEvent struct {}

func (_ FfiDestroyerTypeFungibleVaultWithdrawEvent) Destroy(value FungibleVaultWithdrawEvent) {
	value.Destroy()
}


type IndexedAssertion struct {
	Index uint64
	Assertion Assertion
}

func (r *IndexedAssertion) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.Index);
		FfiDestroyerTypeAssertion{}.Destroy(r.Assertion);
}

type FfiConverterTypeIndexedAssertion struct {}

var FfiConverterTypeIndexedAssertionINSTANCE = FfiConverterTypeIndexedAssertion{}

func (c FfiConverterTypeIndexedAssertion) Lift(rb RustBufferI) IndexedAssertion {
	return LiftFromRustBuffer[IndexedAssertion](c, rb)
}

func (c FfiConverterTypeIndexedAssertion) Read(reader io.Reader) IndexedAssertion {
	return IndexedAssertion {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterTypeAssertionINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeIndexedAssertion) Lower(value IndexedAssertion) RustBuffer {
	return LowerIntoRustBuffer[IndexedAssertion](c, value)
}

func (c FfiConverterTypeIndexedAssertion) Write(writer io.Writer, value IndexedAssertion) {
		FfiConverterUint64INSTANCE.Write(writer, value.Index);
		FfiConverterTypeAssertionINSTANCE.Write(writer, value.Assertion);
}

type FfiDestroyerTypeIndexedAssertion struct {}

func (_ FfiDestroyerTypeIndexedAssertion) Destroy(value IndexedAssertion) {
	value.Destroy()
}


type InitiateBadgeWithdrawAttemptEvent struct {
	Proposer Proposer
}

func (r *InitiateBadgeWithdrawAttemptEvent) Destroy() {
		FfiDestroyerTypeProposer{}.Destroy(r.Proposer);
}

type FfiConverterTypeInitiateBadgeWithdrawAttemptEvent struct {}

var FfiConverterTypeInitiateBadgeWithdrawAttemptEventINSTANCE = FfiConverterTypeInitiateBadgeWithdrawAttemptEvent{}

func (c FfiConverterTypeInitiateBadgeWithdrawAttemptEvent) Lift(rb RustBufferI) InitiateBadgeWithdrawAttemptEvent {
	return LiftFromRustBuffer[InitiateBadgeWithdrawAttemptEvent](c, rb)
}

func (c FfiConverterTypeInitiateBadgeWithdrawAttemptEvent) Read(reader io.Reader) InitiateBadgeWithdrawAttemptEvent {
	return InitiateBadgeWithdrawAttemptEvent {
			FfiConverterTypeProposerINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeInitiateBadgeWithdrawAttemptEvent) Lower(value InitiateBadgeWithdrawAttemptEvent) RustBuffer {
	return LowerIntoRustBuffer[InitiateBadgeWithdrawAttemptEvent](c, value)
}

func (c FfiConverterTypeInitiateBadgeWithdrawAttemptEvent) Write(writer io.Writer, value InitiateBadgeWithdrawAttemptEvent) {
		FfiConverterTypeProposerINSTANCE.Write(writer, value.Proposer);
}

type FfiDestroyerTypeInitiateBadgeWithdrawAttemptEvent struct {}

func (_ FfiDestroyerTypeInitiateBadgeWithdrawAttemptEvent) Destroy(value InitiateBadgeWithdrawAttemptEvent) {
	value.Destroy()
}


type InitiateRecoveryEvent struct {
	Proposer Proposer
	Proposal RecoveryProposal
}

func (r *InitiateRecoveryEvent) Destroy() {
		FfiDestroyerTypeProposer{}.Destroy(r.Proposer);
		FfiDestroyerTypeRecoveryProposal{}.Destroy(r.Proposal);
}

type FfiConverterTypeInitiateRecoveryEvent struct {}

var FfiConverterTypeInitiateRecoveryEventINSTANCE = FfiConverterTypeInitiateRecoveryEvent{}

func (c FfiConverterTypeInitiateRecoveryEvent) Lift(rb RustBufferI) InitiateRecoveryEvent {
	return LiftFromRustBuffer[InitiateRecoveryEvent](c, rb)
}

func (c FfiConverterTypeInitiateRecoveryEvent) Read(reader io.Reader) InitiateRecoveryEvent {
	return InitiateRecoveryEvent {
			FfiConverterTypeProposerINSTANCE.Read(reader),
			FfiConverterTypeRecoveryProposalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeInitiateRecoveryEvent) Lower(value InitiateRecoveryEvent) RustBuffer {
	return LowerIntoRustBuffer[InitiateRecoveryEvent](c, value)
}

func (c FfiConverterTypeInitiateRecoveryEvent) Write(writer io.Writer, value InitiateRecoveryEvent) {
		FfiConverterTypeProposerINSTANCE.Write(writer, value.Proposer);
		FfiConverterTypeRecoveryProposalINSTANCE.Write(writer, value.Proposal);
}

type FfiDestroyerTypeInitiateRecoveryEvent struct {}

func (_ FfiDestroyerTypeInitiateRecoveryEvent) Destroy(value InitiateRecoveryEvent) {
	value.Destroy()
}


type KnownAddresses struct {
	ResourceAddresses ResourceAddresses
	PackageAddresses PackageAddresses
	ComponentAddresses ComponentAddresses
}

func (r *KnownAddresses) Destroy() {
		FfiDestroyerTypeResourceAddresses{}.Destroy(r.ResourceAddresses);
		FfiDestroyerTypePackageAddresses{}.Destroy(r.PackageAddresses);
		FfiDestroyerTypeComponentAddresses{}.Destroy(r.ComponentAddresses);
}

type FfiConverterTypeKnownAddresses struct {}

var FfiConverterTypeKnownAddressesINSTANCE = FfiConverterTypeKnownAddresses{}

func (c FfiConverterTypeKnownAddresses) Lift(rb RustBufferI) KnownAddresses {
	return LiftFromRustBuffer[KnownAddresses](c, rb)
}

func (c FfiConverterTypeKnownAddresses) Read(reader io.Reader) KnownAddresses {
	return KnownAddresses {
			FfiConverterTypeResourceAddressesINSTANCE.Read(reader),
			FfiConverterTypePackageAddressesINSTANCE.Read(reader),
			FfiConverterTypeComponentAddressesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeKnownAddresses) Lower(value KnownAddresses) RustBuffer {
	return LowerIntoRustBuffer[KnownAddresses](c, value)
}

func (c FfiConverterTypeKnownAddresses) Write(writer io.Writer, value KnownAddresses) {
		FfiConverterTypeResourceAddressesINSTANCE.Write(writer, value.ResourceAddresses);
		FfiConverterTypePackageAddressesINSTANCE.Write(writer, value.PackageAddresses);
		FfiConverterTypeComponentAddressesINSTANCE.Write(writer, value.ComponentAddresses);
}

type FfiDestroyerTypeKnownAddresses struct {}

func (_ FfiDestroyerTypeKnownAddresses) Destroy(value KnownAddresses) {
	value.Destroy()
}


type LockFeeModification struct {
	AccountAddress *Address
	Amount *Decimal
}

func (r *LockFeeModification) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.AccountAddress);
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeLockFeeModification struct {}

var FfiConverterTypeLockFeeModificationINSTANCE = FfiConverterTypeLockFeeModification{}

func (c FfiConverterTypeLockFeeModification) Lift(rb RustBufferI) LockFeeModification {
	return LiftFromRustBuffer[LockFeeModification](c, rb)
}

func (c FfiConverterTypeLockFeeModification) Read(reader io.Reader) LockFeeModification {
	return LockFeeModification {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLockFeeModification) Lower(value LockFeeModification) RustBuffer {
	return LowerIntoRustBuffer[LockFeeModification](c, value)
}

func (c FfiConverterTypeLockFeeModification) Write(writer io.Writer, value LockFeeModification) {
		FfiConverterAddressINSTANCE.Write(writer, value.AccountAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeLockFeeModification struct {}

func (_ FfiDestroyerTypeLockFeeModification) Destroy(value LockFeeModification) {
	value.Destroy()
}


type LockOwnerRoleEvent struct {
	PlaceholderField bool
}

func (r *LockOwnerRoleEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.PlaceholderField);
}

type FfiConverterTypeLockOwnerRoleEvent struct {}

var FfiConverterTypeLockOwnerRoleEventINSTANCE = FfiConverterTypeLockOwnerRoleEvent{}

func (c FfiConverterTypeLockOwnerRoleEvent) Lift(rb RustBufferI) LockOwnerRoleEvent {
	return LiftFromRustBuffer[LockOwnerRoleEvent](c, rb)
}

func (c FfiConverterTypeLockOwnerRoleEvent) Read(reader io.Reader) LockOwnerRoleEvent {
	return LockOwnerRoleEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLockOwnerRoleEvent) Lower(value LockOwnerRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[LockOwnerRoleEvent](c, value)
}

func (c FfiConverterTypeLockOwnerRoleEvent) Write(writer io.Writer, value LockOwnerRoleEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.PlaceholderField);
}

type FfiDestroyerTypeLockOwnerRoleEvent struct {}

func (_ FfiDestroyerTypeLockOwnerRoleEvent) Destroy(value LockOwnerRoleEvent) {
	value.Destroy()
}


type LockPrimaryRoleEvent struct {
	PlaceholderField bool
}

func (r *LockPrimaryRoleEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.PlaceholderField);
}

type FfiConverterTypeLockPrimaryRoleEvent struct {}

var FfiConverterTypeLockPrimaryRoleEventINSTANCE = FfiConverterTypeLockPrimaryRoleEvent{}

func (c FfiConverterTypeLockPrimaryRoleEvent) Lift(rb RustBufferI) LockPrimaryRoleEvent {
	return LiftFromRustBuffer[LockPrimaryRoleEvent](c, rb)
}

func (c FfiConverterTypeLockPrimaryRoleEvent) Read(reader io.Reader) LockPrimaryRoleEvent {
	return LockPrimaryRoleEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLockPrimaryRoleEvent) Lower(value LockPrimaryRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[LockPrimaryRoleEvent](c, value)
}

func (c FfiConverterTypeLockPrimaryRoleEvent) Write(writer io.Writer, value LockPrimaryRoleEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.PlaceholderField);
}

type FfiDestroyerTypeLockPrimaryRoleEvent struct {}

func (_ FfiDestroyerTypeLockPrimaryRoleEvent) Destroy(value LockPrimaryRoleEvent) {
	value.Destroy()
}


type LockRoleEvent struct {
	RoleKey string
}

func (r *LockRoleEvent) Destroy() {
		FfiDestroyerString{}.Destroy(r.RoleKey);
}

type FfiConverterTypeLockRoleEvent struct {}

var FfiConverterTypeLockRoleEventINSTANCE = FfiConverterTypeLockRoleEvent{}

func (c FfiConverterTypeLockRoleEvent) Lift(rb RustBufferI) LockRoleEvent {
	return LiftFromRustBuffer[LockRoleEvent](c, rb)
}

func (c FfiConverterTypeLockRoleEvent) Read(reader io.Reader) LockRoleEvent {
	return LockRoleEvent {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLockRoleEvent) Lower(value LockRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[LockRoleEvent](c, value)
}

func (c FfiConverterTypeLockRoleEvent) Write(writer io.Writer, value LockRoleEvent) {
		FfiConverterStringINSTANCE.Write(writer, value.RoleKey);
}

type FfiDestroyerTypeLockRoleEvent struct {}

func (_ FfiDestroyerTypeLockRoleEvent) Destroy(value LockRoleEvent) {
	value.Destroy()
}


type ManifestAddressReservation struct {
	Value uint32
}

func (r *ManifestAddressReservation) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.Value);
}

type FfiConverterTypeManifestAddressReservation struct {}

var FfiConverterTypeManifestAddressReservationINSTANCE = FfiConverterTypeManifestAddressReservation{}

func (c FfiConverterTypeManifestAddressReservation) Lift(rb RustBufferI) ManifestAddressReservation {
	return LiftFromRustBuffer[ManifestAddressReservation](c, rb)
}

func (c FfiConverterTypeManifestAddressReservation) Read(reader io.Reader) ManifestAddressReservation {
	return ManifestAddressReservation {
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestAddressReservation) Lower(value ManifestAddressReservation) RustBuffer {
	return LowerIntoRustBuffer[ManifestAddressReservation](c, value)
}

func (c FfiConverterTypeManifestAddressReservation) Write(writer io.Writer, value ManifestAddressReservation) {
		FfiConverterUint32INSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeManifestAddressReservation struct {}

func (_ FfiDestroyerTypeManifestAddressReservation) Destroy(value ManifestAddressReservation) {
	value.Destroy()
}


type ManifestBlobRef struct {
	Value *Hash
}

func (r *ManifestBlobRef) Destroy() {
		FfiDestroyerHash{}.Destroy(r.Value);
}

type FfiConverterTypeManifestBlobRef struct {}

var FfiConverterTypeManifestBlobRefINSTANCE = FfiConverterTypeManifestBlobRef{}

func (c FfiConverterTypeManifestBlobRef) Lift(rb RustBufferI) ManifestBlobRef {
	return LiftFromRustBuffer[ManifestBlobRef](c, rb)
}

func (c FfiConverterTypeManifestBlobRef) Read(reader io.Reader) ManifestBlobRef {
	return ManifestBlobRef {
			FfiConverterHashINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBlobRef) Lower(value ManifestBlobRef) RustBuffer {
	return LowerIntoRustBuffer[ManifestBlobRef](c, value)
}

func (c FfiConverterTypeManifestBlobRef) Write(writer io.Writer, value ManifestBlobRef) {
		FfiConverterHashINSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeManifestBlobRef struct {}

func (_ FfiDestroyerTypeManifestBlobRef) Destroy(value ManifestBlobRef) {
	value.Destroy()
}


type ManifestBucket struct {
	Value uint32
}

func (r *ManifestBucket) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.Value);
}

type FfiConverterTypeManifestBucket struct {}

var FfiConverterTypeManifestBucketINSTANCE = FfiConverterTypeManifestBucket{}

func (c FfiConverterTypeManifestBucket) Lift(rb RustBufferI) ManifestBucket {
	return LiftFromRustBuffer[ManifestBucket](c, rb)
}

func (c FfiConverterTypeManifestBucket) Read(reader io.Reader) ManifestBucket {
	return ManifestBucket {
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBucket) Lower(value ManifestBucket) RustBuffer {
	return LowerIntoRustBuffer[ManifestBucket](c, value)
}

func (c FfiConverterTypeManifestBucket) Write(writer io.Writer, value ManifestBucket) {
		FfiConverterUint32INSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeManifestBucket struct {}

func (_ FfiDestroyerTypeManifestBucket) Destroy(value ManifestBucket) {
	value.Destroy()
}


type ManifestBuilderAddressReservation struct {
	Name string
}

func (r *ManifestBuilderAddressReservation) Destroy() {
		FfiDestroyerString{}.Destroy(r.Name);
}

type FfiConverterTypeManifestBuilderAddressReservation struct {}

var FfiConverterTypeManifestBuilderAddressReservationINSTANCE = FfiConverterTypeManifestBuilderAddressReservation{}

func (c FfiConverterTypeManifestBuilderAddressReservation) Lift(rb RustBufferI) ManifestBuilderAddressReservation {
	return LiftFromRustBuffer[ManifestBuilderAddressReservation](c, rb)
}

func (c FfiConverterTypeManifestBuilderAddressReservation) Read(reader io.Reader) ManifestBuilderAddressReservation {
	return ManifestBuilderAddressReservation {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBuilderAddressReservation) Lower(value ManifestBuilderAddressReservation) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderAddressReservation](c, value)
}

func (c FfiConverterTypeManifestBuilderAddressReservation) Write(writer io.Writer, value ManifestBuilderAddressReservation) {
		FfiConverterStringINSTANCE.Write(writer, value.Name);
}

type FfiDestroyerTypeManifestBuilderAddressReservation struct {}

func (_ FfiDestroyerTypeManifestBuilderAddressReservation) Destroy(value ManifestBuilderAddressReservation) {
	value.Destroy()
}


type ManifestBuilderBucket struct {
	Name string
}

func (r *ManifestBuilderBucket) Destroy() {
		FfiDestroyerString{}.Destroy(r.Name);
}

type FfiConverterTypeManifestBuilderBucket struct {}

var FfiConverterTypeManifestBuilderBucketINSTANCE = FfiConverterTypeManifestBuilderBucket{}

func (c FfiConverterTypeManifestBuilderBucket) Lift(rb RustBufferI) ManifestBuilderBucket {
	return LiftFromRustBuffer[ManifestBuilderBucket](c, rb)
}

func (c FfiConverterTypeManifestBuilderBucket) Read(reader io.Reader) ManifestBuilderBucket {
	return ManifestBuilderBucket {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBuilderBucket) Lower(value ManifestBuilderBucket) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderBucket](c, value)
}

func (c FfiConverterTypeManifestBuilderBucket) Write(writer io.Writer, value ManifestBuilderBucket) {
		FfiConverterStringINSTANCE.Write(writer, value.Name);
}

type FfiDestroyerTypeManifestBuilderBucket struct {}

func (_ FfiDestroyerTypeManifestBuilderBucket) Destroy(value ManifestBuilderBucket) {
	value.Destroy()
}


type ManifestBuilderMapEntry struct {
	Key ManifestBuilderValue
	Value ManifestBuilderValue
}

func (r *ManifestBuilderMapEntry) Destroy() {
		FfiDestroyerTypeManifestBuilderValue{}.Destroy(r.Key);
		FfiDestroyerTypeManifestBuilderValue{}.Destroy(r.Value);
}

type FfiConverterTypeManifestBuilderMapEntry struct {}

var FfiConverterTypeManifestBuilderMapEntryINSTANCE = FfiConverterTypeManifestBuilderMapEntry{}

func (c FfiConverterTypeManifestBuilderMapEntry) Lift(rb RustBufferI) ManifestBuilderMapEntry {
	return LiftFromRustBuffer[ManifestBuilderMapEntry](c, rb)
}

func (c FfiConverterTypeManifestBuilderMapEntry) Read(reader io.Reader) ManifestBuilderMapEntry {
	return ManifestBuilderMapEntry {
			FfiConverterTypeManifestBuilderValueINSTANCE.Read(reader),
			FfiConverterTypeManifestBuilderValueINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBuilderMapEntry) Lower(value ManifestBuilderMapEntry) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderMapEntry](c, value)
}

func (c FfiConverterTypeManifestBuilderMapEntry) Write(writer io.Writer, value ManifestBuilderMapEntry) {
		FfiConverterTypeManifestBuilderValueINSTANCE.Write(writer, value.Key);
		FfiConverterTypeManifestBuilderValueINSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeManifestBuilderMapEntry struct {}

func (_ FfiDestroyerTypeManifestBuilderMapEntry) Destroy(value ManifestBuilderMapEntry) {
	value.Destroy()
}


type ManifestBuilderNamedAddress struct {
	Name string
}

func (r *ManifestBuilderNamedAddress) Destroy() {
		FfiDestroyerString{}.Destroy(r.Name);
}

type FfiConverterTypeManifestBuilderNamedAddress struct {}

var FfiConverterTypeManifestBuilderNamedAddressINSTANCE = FfiConverterTypeManifestBuilderNamedAddress{}

func (c FfiConverterTypeManifestBuilderNamedAddress) Lift(rb RustBufferI) ManifestBuilderNamedAddress {
	return LiftFromRustBuffer[ManifestBuilderNamedAddress](c, rb)
}

func (c FfiConverterTypeManifestBuilderNamedAddress) Read(reader io.Reader) ManifestBuilderNamedAddress {
	return ManifestBuilderNamedAddress {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBuilderNamedAddress) Lower(value ManifestBuilderNamedAddress) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderNamedAddress](c, value)
}

func (c FfiConverterTypeManifestBuilderNamedAddress) Write(writer io.Writer, value ManifestBuilderNamedAddress) {
		FfiConverterStringINSTANCE.Write(writer, value.Name);
}

type FfiDestroyerTypeManifestBuilderNamedAddress struct {}

func (_ FfiDestroyerTypeManifestBuilderNamedAddress) Destroy(value ManifestBuilderNamedAddress) {
	value.Destroy()
}


type ManifestBuilderProof struct {
	Name string
}

func (r *ManifestBuilderProof) Destroy() {
		FfiDestroyerString{}.Destroy(r.Name);
}

type FfiConverterTypeManifestBuilderProof struct {}

var FfiConverterTypeManifestBuilderProofINSTANCE = FfiConverterTypeManifestBuilderProof{}

func (c FfiConverterTypeManifestBuilderProof) Lift(rb RustBufferI) ManifestBuilderProof {
	return LiftFromRustBuffer[ManifestBuilderProof](c, rb)
}

func (c FfiConverterTypeManifestBuilderProof) Read(reader io.Reader) ManifestBuilderProof {
	return ManifestBuilderProof {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestBuilderProof) Lower(value ManifestBuilderProof) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderProof](c, value)
}

func (c FfiConverterTypeManifestBuilderProof) Write(writer io.Writer, value ManifestBuilderProof) {
		FfiConverterStringINSTANCE.Write(writer, value.Name);
}

type FfiDestroyerTypeManifestBuilderProof struct {}

func (_ FfiDestroyerTypeManifestBuilderProof) Destroy(value ManifestBuilderProof) {
	value.Destroy()
}


type ManifestProof struct {
	Value uint32
}

func (r *ManifestProof) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.Value);
}

type FfiConverterTypeManifestProof struct {}

var FfiConverterTypeManifestProofINSTANCE = FfiConverterTypeManifestProof{}

func (c FfiConverterTypeManifestProof) Lift(rb RustBufferI) ManifestProof {
	return LiftFromRustBuffer[ManifestProof](c, rb)
}

func (c FfiConverterTypeManifestProof) Read(reader io.Reader) ManifestProof {
	return ManifestProof {
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestProof) Lower(value ManifestProof) RustBuffer {
	return LowerIntoRustBuffer[ManifestProof](c, value)
}

func (c FfiConverterTypeManifestProof) Write(writer io.Writer, value ManifestProof) {
		FfiConverterUint32INSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeManifestProof struct {}

func (_ FfiDestroyerTypeManifestProof) Destroy(value ManifestProof) {
	value.Destroy()
}


type ManifestSummary struct {
	PresentedProofs map[string][]ResourceSpecifier
	AccountsWithdrawnFrom []*Address
	AccountsDepositedInto []*Address
	EncounteredEntities []*Address
	AccountsRequiringAuth []*Address
	IdentitiesRequiringAuth []*Address
	ReservedInstructions []ReservedInstruction
	Classification []ManifestClass
}

func (r *ManifestSummary) Destroy() {
		FfiDestroyerMapStringSequenceTypeResourceSpecifier{}.Destroy(r.PresentedProofs);
		FfiDestroyerSequenceAddress{}.Destroy(r.AccountsWithdrawnFrom);
		FfiDestroyerSequenceAddress{}.Destroy(r.AccountsDepositedInto);
		FfiDestroyerSequenceAddress{}.Destroy(r.EncounteredEntities);
		FfiDestroyerSequenceAddress{}.Destroy(r.AccountsRequiringAuth);
		FfiDestroyerSequenceAddress{}.Destroy(r.IdentitiesRequiringAuth);
		FfiDestroyerSequenceTypeReservedInstruction{}.Destroy(r.ReservedInstructions);
		FfiDestroyerSequenceTypeManifestClass{}.Destroy(r.Classification);
}

type FfiConverterTypeManifestSummary struct {}

var FfiConverterTypeManifestSummaryINSTANCE = FfiConverterTypeManifestSummary{}

func (c FfiConverterTypeManifestSummary) Lift(rb RustBufferI) ManifestSummary {
	return LiftFromRustBuffer[ManifestSummary](c, rb)
}

func (c FfiConverterTypeManifestSummary) Read(reader io.Reader) ManifestSummary {
	return ManifestSummary {
			FfiConverterMapStringSequenceTypeResourceSpecifierINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceTypeReservedInstructionINSTANCE.Read(reader),
			FfiConverterSequenceTypeManifestClassINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeManifestSummary) Lower(value ManifestSummary) RustBuffer {
	return LowerIntoRustBuffer[ManifestSummary](c, value)
}

func (c FfiConverterTypeManifestSummary) Write(writer io.Writer, value ManifestSummary) {
		FfiConverterMapStringSequenceTypeResourceSpecifierINSTANCE.Write(writer, value.PresentedProofs);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.AccountsWithdrawnFrom);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.AccountsDepositedInto);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.EncounteredEntities);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.AccountsRequiringAuth);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.IdentitiesRequiringAuth);
		FfiConverterSequenceTypeReservedInstructionINSTANCE.Write(writer, value.ReservedInstructions);
		FfiConverterSequenceTypeManifestClassINSTANCE.Write(writer, value.Classification);
}

type FfiDestroyerTypeManifestSummary struct {}

func (_ FfiDestroyerTypeManifestSummary) Destroy(value ManifestSummary) {
	value.Destroy()
}


type MapEntry struct {
	Key ManifestValue
	Value ManifestValue
}

func (r *MapEntry) Destroy() {
		FfiDestroyerTypeManifestValue{}.Destroy(r.Key);
		FfiDestroyerTypeManifestValue{}.Destroy(r.Value);
}

type FfiConverterTypeMapEntry struct {}

var FfiConverterTypeMapEntryINSTANCE = FfiConverterTypeMapEntry{}

func (c FfiConverterTypeMapEntry) Lift(rb RustBufferI) MapEntry {
	return LiftFromRustBuffer[MapEntry](c, rb)
}

func (c FfiConverterTypeMapEntry) Read(reader io.Reader) MapEntry {
	return MapEntry {
			FfiConverterTypeManifestValueINSTANCE.Read(reader),
			FfiConverterTypeManifestValueINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMapEntry) Lower(value MapEntry) RustBuffer {
	return LowerIntoRustBuffer[MapEntry](c, value)
}

func (c FfiConverterTypeMapEntry) Write(writer io.Writer, value MapEntry) {
		FfiConverterTypeManifestValueINSTANCE.Write(writer, value.Key);
		FfiConverterTypeManifestValueINSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeMapEntry struct {}

func (_ FfiDestroyerTypeMapEntry) Destroy(value MapEntry) {
	value.Destroy()
}


type MetadataInitEntry struct {
	Value *MetadataValue
	Lock bool
}

func (r *MetadataInitEntry) Destroy() {
		FfiDestroyerOptionalTypeMetadataValue{}.Destroy(r.Value);
		FfiDestroyerBool{}.Destroy(r.Lock);
}

type FfiConverterTypeMetadataInitEntry struct {}

var FfiConverterTypeMetadataInitEntryINSTANCE = FfiConverterTypeMetadataInitEntry{}

func (c FfiConverterTypeMetadataInitEntry) Lift(rb RustBufferI) MetadataInitEntry {
	return LiftFromRustBuffer[MetadataInitEntry](c, rb)
}

func (c FfiConverterTypeMetadataInitEntry) Read(reader io.Reader) MetadataInitEntry {
	return MetadataInitEntry {
			FfiConverterOptionalTypeMetadataValueINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMetadataInitEntry) Lower(value MetadataInitEntry) RustBuffer {
	return LowerIntoRustBuffer[MetadataInitEntry](c, value)
}

func (c FfiConverterTypeMetadataInitEntry) Write(writer io.Writer, value MetadataInitEntry) {
		FfiConverterOptionalTypeMetadataValueINSTANCE.Write(writer, value.Value);
		FfiConverterBoolINSTANCE.Write(writer, value.Lock);
}

type FfiDestroyerTypeMetadataInitEntry struct {}

func (_ FfiDestroyerTypeMetadataInitEntry) Destroy(value MetadataInitEntry) {
	value.Destroy()
}


type MetadataModuleConfig struct {
	Init map[string]MetadataInitEntry
	Roles map[string]**AccessRule
}

func (r *MetadataModuleConfig) Destroy() {
		FfiDestroyerMapStringTypeMetadataInitEntry{}.Destroy(r.Init);
		FfiDestroyerMapStringOptionalAccessRule{}.Destroy(r.Roles);
}

type FfiConverterTypeMetadataModuleConfig struct {}

var FfiConverterTypeMetadataModuleConfigINSTANCE = FfiConverterTypeMetadataModuleConfig{}

func (c FfiConverterTypeMetadataModuleConfig) Lift(rb RustBufferI) MetadataModuleConfig {
	return LiftFromRustBuffer[MetadataModuleConfig](c, rb)
}

func (c FfiConverterTypeMetadataModuleConfig) Read(reader io.Reader) MetadataModuleConfig {
	return MetadataModuleConfig {
			FfiConverterMapStringTypeMetadataInitEntryINSTANCE.Read(reader),
			FfiConverterMapStringOptionalAccessRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMetadataModuleConfig) Lower(value MetadataModuleConfig) RustBuffer {
	return LowerIntoRustBuffer[MetadataModuleConfig](c, value)
}

func (c FfiConverterTypeMetadataModuleConfig) Write(writer io.Writer, value MetadataModuleConfig) {
		FfiConverterMapStringTypeMetadataInitEntryINSTANCE.Write(writer, value.Init);
		FfiConverterMapStringOptionalAccessRuleINSTANCE.Write(writer, value.Roles);
}

type FfiDestroyerTypeMetadataModuleConfig struct {}

func (_ FfiDestroyerTypeMetadataModuleConfig) Destroy(value MetadataModuleConfig) {
	value.Destroy()
}


type MintFungibleResourceEvent struct {
	Amount *Decimal
}

func (r *MintFungibleResourceEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeMintFungibleResourceEvent struct {}

var FfiConverterTypeMintFungibleResourceEventINSTANCE = FfiConverterTypeMintFungibleResourceEvent{}

func (c FfiConverterTypeMintFungibleResourceEvent) Lift(rb RustBufferI) MintFungibleResourceEvent {
	return LiftFromRustBuffer[MintFungibleResourceEvent](c, rb)
}

func (c FfiConverterTypeMintFungibleResourceEvent) Read(reader io.Reader) MintFungibleResourceEvent {
	return MintFungibleResourceEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMintFungibleResourceEvent) Lower(value MintFungibleResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[MintFungibleResourceEvent](c, value)
}

func (c FfiConverterTypeMintFungibleResourceEvent) Write(writer io.Writer, value MintFungibleResourceEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeMintFungibleResourceEvent struct {}

func (_ FfiDestroyerTypeMintFungibleResourceEvent) Destroy(value MintFungibleResourceEvent) {
	value.Destroy()
}


type MintNonFungibleResourceEvent struct {
	Ids []NonFungibleLocalId
}

func (r *MintNonFungibleResourceEvent) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.Ids);
}

type FfiConverterTypeMintNonFungibleResourceEvent struct {}

var FfiConverterTypeMintNonFungibleResourceEventINSTANCE = FfiConverterTypeMintNonFungibleResourceEvent{}

func (c FfiConverterTypeMintNonFungibleResourceEvent) Lift(rb RustBufferI) MintNonFungibleResourceEvent {
	return LiftFromRustBuffer[MintNonFungibleResourceEvent](c, rb)
}

func (c FfiConverterTypeMintNonFungibleResourceEvent) Read(reader io.Reader) MintNonFungibleResourceEvent {
	return MintNonFungibleResourceEvent {
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMintNonFungibleResourceEvent) Lower(value MintNonFungibleResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[MintNonFungibleResourceEvent](c, value)
}

func (c FfiConverterTypeMintNonFungibleResourceEvent) Write(writer io.Writer, value MintNonFungibleResourceEvent) {
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.Ids);
}

type FfiDestroyerTypeMintNonFungibleResourceEvent struct {}

func (_ FfiDestroyerTypeMintNonFungibleResourceEvent) Destroy(value MintNonFungibleResourceEvent) {
	value.Destroy()
}


type MultiResourcePoolContributionEvent struct {
	ContributedResources map[string]*Decimal
	PoolUnitsMinted *Decimal
}

func (r *MultiResourcePoolContributionEvent) Destroy() {
		FfiDestroyerMapStringDecimal{}.Destroy(r.ContributedResources);
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitsMinted);
}

type FfiConverterTypeMultiResourcePoolContributionEvent struct {}

var FfiConverterTypeMultiResourcePoolContributionEventINSTANCE = FfiConverterTypeMultiResourcePoolContributionEvent{}

func (c FfiConverterTypeMultiResourcePoolContributionEvent) Lift(rb RustBufferI) MultiResourcePoolContributionEvent {
	return LiftFromRustBuffer[MultiResourcePoolContributionEvent](c, rb)
}

func (c FfiConverterTypeMultiResourcePoolContributionEvent) Read(reader io.Reader) MultiResourcePoolContributionEvent {
	return MultiResourcePoolContributionEvent {
			FfiConverterMapStringDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMultiResourcePoolContributionEvent) Lower(value MultiResourcePoolContributionEvent) RustBuffer {
	return LowerIntoRustBuffer[MultiResourcePoolContributionEvent](c, value)
}

func (c FfiConverterTypeMultiResourcePoolContributionEvent) Write(writer io.Writer, value MultiResourcePoolContributionEvent) {
		FfiConverterMapStringDecimalINSTANCE.Write(writer, value.ContributedResources);
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitsMinted);
}

type FfiDestroyerTypeMultiResourcePoolContributionEvent struct {}

func (_ FfiDestroyerTypeMultiResourcePoolContributionEvent) Destroy(value MultiResourcePoolContributionEvent) {
	value.Destroy()
}


type MultiResourcePoolDepositEvent struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (r *MultiResourcePoolDepositEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeMultiResourcePoolDepositEvent struct {}

var FfiConverterTypeMultiResourcePoolDepositEventINSTANCE = FfiConverterTypeMultiResourcePoolDepositEvent{}

func (c FfiConverterTypeMultiResourcePoolDepositEvent) Lift(rb RustBufferI) MultiResourcePoolDepositEvent {
	return LiftFromRustBuffer[MultiResourcePoolDepositEvent](c, rb)
}

func (c FfiConverterTypeMultiResourcePoolDepositEvent) Read(reader io.Reader) MultiResourcePoolDepositEvent {
	return MultiResourcePoolDepositEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMultiResourcePoolDepositEvent) Lower(value MultiResourcePoolDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[MultiResourcePoolDepositEvent](c, value)
}

func (c FfiConverterTypeMultiResourcePoolDepositEvent) Write(writer io.Writer, value MultiResourcePoolDepositEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeMultiResourcePoolDepositEvent struct {}

func (_ FfiDestroyerTypeMultiResourcePoolDepositEvent) Destroy(value MultiResourcePoolDepositEvent) {
	value.Destroy()
}


type MultiResourcePoolRedemptionEvent struct {
	PoolUnitTokensRedeemed *Decimal
	RedeemedResources map[string]*Decimal
}

func (r *MultiResourcePoolRedemptionEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitTokensRedeemed);
		FfiDestroyerMapStringDecimal{}.Destroy(r.RedeemedResources);
}

type FfiConverterTypeMultiResourcePoolRedemptionEvent struct {}

var FfiConverterTypeMultiResourcePoolRedemptionEventINSTANCE = FfiConverterTypeMultiResourcePoolRedemptionEvent{}

func (c FfiConverterTypeMultiResourcePoolRedemptionEvent) Lift(rb RustBufferI) MultiResourcePoolRedemptionEvent {
	return LiftFromRustBuffer[MultiResourcePoolRedemptionEvent](c, rb)
}

func (c FfiConverterTypeMultiResourcePoolRedemptionEvent) Read(reader io.Reader) MultiResourcePoolRedemptionEvent {
	return MultiResourcePoolRedemptionEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterMapStringDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMultiResourcePoolRedemptionEvent) Lower(value MultiResourcePoolRedemptionEvent) RustBuffer {
	return LowerIntoRustBuffer[MultiResourcePoolRedemptionEvent](c, value)
}

func (c FfiConverterTypeMultiResourcePoolRedemptionEvent) Write(writer io.Writer, value MultiResourcePoolRedemptionEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitTokensRedeemed);
		FfiConverterMapStringDecimalINSTANCE.Write(writer, value.RedeemedResources);
}

type FfiDestroyerTypeMultiResourcePoolRedemptionEvent struct {}

func (_ FfiDestroyerTypeMultiResourcePoolRedemptionEvent) Destroy(value MultiResourcePoolRedemptionEvent) {
	value.Destroy()
}


type MultiResourcePoolWithdrawEvent struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (r *MultiResourcePoolWithdrawEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeMultiResourcePoolWithdrawEvent struct {}

var FfiConverterTypeMultiResourcePoolWithdrawEventINSTANCE = FfiConverterTypeMultiResourcePoolWithdrawEvent{}

func (c FfiConverterTypeMultiResourcePoolWithdrawEvent) Lift(rb RustBufferI) MultiResourcePoolWithdrawEvent {
	return LiftFromRustBuffer[MultiResourcePoolWithdrawEvent](c, rb)
}

func (c FfiConverterTypeMultiResourcePoolWithdrawEvent) Read(reader io.Reader) MultiResourcePoolWithdrawEvent {
	return MultiResourcePoolWithdrawEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMultiResourcePoolWithdrawEvent) Lower(value MultiResourcePoolWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[MultiResourcePoolWithdrawEvent](c, value)
}

func (c FfiConverterTypeMultiResourcePoolWithdrawEvent) Write(writer io.Writer, value MultiResourcePoolWithdrawEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeMultiResourcePoolWithdrawEvent struct {}

func (_ FfiDestroyerTypeMultiResourcePoolWithdrawEvent) Destroy(value MultiResourcePoolWithdrawEvent) {
	value.Destroy()
}


type NewEntities struct {
	ComponentAddresses []*Address
	ResourceAddresses []*Address
	PackageAddresses []*Address
	Metadata map[string]map[string]*MetadataValue
}

func (r *NewEntities) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(r.ComponentAddresses);
		FfiDestroyerSequenceAddress{}.Destroy(r.ResourceAddresses);
		FfiDestroyerSequenceAddress{}.Destroy(r.PackageAddresses);
		FfiDestroyerMapStringMapStringOptionalTypeMetadataValue{}.Destroy(r.Metadata);
}

type FfiConverterTypeNewEntities struct {}

var FfiConverterTypeNewEntitiesINSTANCE = FfiConverterTypeNewEntities{}

func (c FfiConverterTypeNewEntities) Lift(rb RustBufferI) NewEntities {
	return LiftFromRustBuffer[NewEntities](c, rb)
}

func (c FfiConverterTypeNewEntities) Read(reader io.Reader) NewEntities {
	return NewEntities {
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterMapStringMapStringOptionalTypeMetadataValueINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNewEntities) Lower(value NewEntities) RustBuffer {
	return LowerIntoRustBuffer[NewEntities](c, value)
}

func (c FfiConverterTypeNewEntities) Write(writer io.Writer, value NewEntities) {
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.ComponentAddresses);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.ResourceAddresses);
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.PackageAddresses);
		FfiConverterMapStringMapStringOptionalTypeMetadataValueINSTANCE.Write(writer, value.Metadata);
}

type FfiDestroyerTypeNewEntities struct {}

func (_ FfiDestroyerTypeNewEntities) Destroy(value NewEntities) {
	value.Destroy()
}


type NonFungibleVaultDepositEvent struct {
	Ids []NonFungibleLocalId
}

func (r *NonFungibleVaultDepositEvent) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.Ids);
}

type FfiConverterTypeNonFungibleVaultDepositEvent struct {}

var FfiConverterTypeNonFungibleVaultDepositEventINSTANCE = FfiConverterTypeNonFungibleVaultDepositEvent{}

func (c FfiConverterTypeNonFungibleVaultDepositEvent) Lift(rb RustBufferI) NonFungibleVaultDepositEvent {
	return LiftFromRustBuffer[NonFungibleVaultDepositEvent](c, rb)
}

func (c FfiConverterTypeNonFungibleVaultDepositEvent) Read(reader io.Reader) NonFungibleVaultDepositEvent {
	return NonFungibleVaultDepositEvent {
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNonFungibleVaultDepositEvent) Lower(value NonFungibleVaultDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[NonFungibleVaultDepositEvent](c, value)
}

func (c FfiConverterTypeNonFungibleVaultDepositEvent) Write(writer io.Writer, value NonFungibleVaultDepositEvent) {
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.Ids);
}

type FfiDestroyerTypeNonFungibleVaultDepositEvent struct {}

func (_ FfiDestroyerTypeNonFungibleVaultDepositEvent) Destroy(value NonFungibleVaultDepositEvent) {
	value.Destroy()
}


type NonFungibleVaultRecallEvent struct {
	Ids []NonFungibleLocalId
}

func (r *NonFungibleVaultRecallEvent) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.Ids);
}

type FfiConverterTypeNonFungibleVaultRecallEvent struct {}

var FfiConverterTypeNonFungibleVaultRecallEventINSTANCE = FfiConverterTypeNonFungibleVaultRecallEvent{}

func (c FfiConverterTypeNonFungibleVaultRecallEvent) Lift(rb RustBufferI) NonFungibleVaultRecallEvent {
	return LiftFromRustBuffer[NonFungibleVaultRecallEvent](c, rb)
}

func (c FfiConverterTypeNonFungibleVaultRecallEvent) Read(reader io.Reader) NonFungibleVaultRecallEvent {
	return NonFungibleVaultRecallEvent {
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNonFungibleVaultRecallEvent) Lower(value NonFungibleVaultRecallEvent) RustBuffer {
	return LowerIntoRustBuffer[NonFungibleVaultRecallEvent](c, value)
}

func (c FfiConverterTypeNonFungibleVaultRecallEvent) Write(writer io.Writer, value NonFungibleVaultRecallEvent) {
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.Ids);
}

type FfiDestroyerTypeNonFungibleVaultRecallEvent struct {}

func (_ FfiDestroyerTypeNonFungibleVaultRecallEvent) Destroy(value NonFungibleVaultRecallEvent) {
	value.Destroy()
}


type NonFungibleVaultWithdrawEvent struct {
	Ids []NonFungibleLocalId
}

func (r *NonFungibleVaultWithdrawEvent) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.Ids);
}

type FfiConverterTypeNonFungibleVaultWithdrawEvent struct {}

var FfiConverterTypeNonFungibleVaultWithdrawEventINSTANCE = FfiConverterTypeNonFungibleVaultWithdrawEvent{}

func (c FfiConverterTypeNonFungibleVaultWithdrawEvent) Lift(rb RustBufferI) NonFungibleVaultWithdrawEvent {
	return LiftFromRustBuffer[NonFungibleVaultWithdrawEvent](c, rb)
}

func (c FfiConverterTypeNonFungibleVaultWithdrawEvent) Read(reader io.Reader) NonFungibleVaultWithdrawEvent {
	return NonFungibleVaultWithdrawEvent {
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNonFungibleVaultWithdrawEvent) Lower(value NonFungibleVaultWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[NonFungibleVaultWithdrawEvent](c, value)
}

func (c FfiConverterTypeNonFungibleVaultWithdrawEvent) Write(writer io.Writer, value NonFungibleVaultWithdrawEvent) {
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.Ids);
}

type FfiDestroyerTypeNonFungibleVaultWithdrawEvent struct {}

func (_ FfiDestroyerTypeNonFungibleVaultWithdrawEvent) Destroy(value NonFungibleVaultWithdrawEvent) {
	value.Destroy()
}


type OneResourcePoolContributionEvent struct {
	AmountOfResourcesContributed *Decimal
	PoolUnitsMinted *Decimal
}

func (r *OneResourcePoolContributionEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.AmountOfResourcesContributed);
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitsMinted);
}

type FfiConverterTypeOneResourcePoolContributionEvent struct {}

var FfiConverterTypeOneResourcePoolContributionEventINSTANCE = FfiConverterTypeOneResourcePoolContributionEvent{}

func (c FfiConverterTypeOneResourcePoolContributionEvent) Lift(rb RustBufferI) OneResourcePoolContributionEvent {
	return LiftFromRustBuffer[OneResourcePoolContributionEvent](c, rb)
}

func (c FfiConverterTypeOneResourcePoolContributionEvent) Read(reader io.Reader) OneResourcePoolContributionEvent {
	return OneResourcePoolContributionEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOneResourcePoolContributionEvent) Lower(value OneResourcePoolContributionEvent) RustBuffer {
	return LowerIntoRustBuffer[OneResourcePoolContributionEvent](c, value)
}

func (c FfiConverterTypeOneResourcePoolContributionEvent) Write(writer io.Writer, value OneResourcePoolContributionEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.AmountOfResourcesContributed);
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitsMinted);
}

type FfiDestroyerTypeOneResourcePoolContributionEvent struct {}

func (_ FfiDestroyerTypeOneResourcePoolContributionEvent) Destroy(value OneResourcePoolContributionEvent) {
	value.Destroy()
}


type OneResourcePoolDepositEvent struct {
	Amount *Decimal
}

func (r *OneResourcePoolDepositEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeOneResourcePoolDepositEvent struct {}

var FfiConverterTypeOneResourcePoolDepositEventINSTANCE = FfiConverterTypeOneResourcePoolDepositEvent{}

func (c FfiConverterTypeOneResourcePoolDepositEvent) Lift(rb RustBufferI) OneResourcePoolDepositEvent {
	return LiftFromRustBuffer[OneResourcePoolDepositEvent](c, rb)
}

func (c FfiConverterTypeOneResourcePoolDepositEvent) Read(reader io.Reader) OneResourcePoolDepositEvent {
	return OneResourcePoolDepositEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOneResourcePoolDepositEvent) Lower(value OneResourcePoolDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[OneResourcePoolDepositEvent](c, value)
}

func (c FfiConverterTypeOneResourcePoolDepositEvent) Write(writer io.Writer, value OneResourcePoolDepositEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeOneResourcePoolDepositEvent struct {}

func (_ FfiDestroyerTypeOneResourcePoolDepositEvent) Destroy(value OneResourcePoolDepositEvent) {
	value.Destroy()
}


type OneResourcePoolRedemptionEvent struct {
	PoolUnitTokensRedeemed *Decimal
	RedeemedAmount *Decimal
}

func (r *OneResourcePoolRedemptionEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitTokensRedeemed);
		FfiDestroyerDecimal{}.Destroy(r.RedeemedAmount);
}

type FfiConverterTypeOneResourcePoolRedemptionEvent struct {}

var FfiConverterTypeOneResourcePoolRedemptionEventINSTANCE = FfiConverterTypeOneResourcePoolRedemptionEvent{}

func (c FfiConverterTypeOneResourcePoolRedemptionEvent) Lift(rb RustBufferI) OneResourcePoolRedemptionEvent {
	return LiftFromRustBuffer[OneResourcePoolRedemptionEvent](c, rb)
}

func (c FfiConverterTypeOneResourcePoolRedemptionEvent) Read(reader io.Reader) OneResourcePoolRedemptionEvent {
	return OneResourcePoolRedemptionEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOneResourcePoolRedemptionEvent) Lower(value OneResourcePoolRedemptionEvent) RustBuffer {
	return LowerIntoRustBuffer[OneResourcePoolRedemptionEvent](c, value)
}

func (c FfiConverterTypeOneResourcePoolRedemptionEvent) Write(writer io.Writer, value OneResourcePoolRedemptionEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitTokensRedeemed);
		FfiConverterDecimalINSTANCE.Write(writer, value.RedeemedAmount);
}

type FfiDestroyerTypeOneResourcePoolRedemptionEvent struct {}

func (_ FfiDestroyerTypeOneResourcePoolRedemptionEvent) Destroy(value OneResourcePoolRedemptionEvent) {
	value.Destroy()
}


type OneResourcePoolWithdrawEvent struct {
	Amount *Decimal
}

func (r *OneResourcePoolWithdrawEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeOneResourcePoolWithdrawEvent struct {}

var FfiConverterTypeOneResourcePoolWithdrawEventINSTANCE = FfiConverterTypeOneResourcePoolWithdrawEvent{}

func (c FfiConverterTypeOneResourcePoolWithdrawEvent) Lift(rb RustBufferI) OneResourcePoolWithdrawEvent {
	return LiftFromRustBuffer[OneResourcePoolWithdrawEvent](c, rb)
}

func (c FfiConverterTypeOneResourcePoolWithdrawEvent) Read(reader io.Reader) OneResourcePoolWithdrawEvent {
	return OneResourcePoolWithdrawEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOneResourcePoolWithdrawEvent) Lower(value OneResourcePoolWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[OneResourcePoolWithdrawEvent](c, value)
}

func (c FfiConverterTypeOneResourcePoolWithdrawEvent) Write(writer io.Writer, value OneResourcePoolWithdrawEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeOneResourcePoolWithdrawEvent struct {}

func (_ FfiDestroyerTypeOneResourcePoolWithdrawEvent) Destroy(value OneResourcePoolWithdrawEvent) {
	value.Destroy()
}


type PackageAddresses struct {
	PackagePackage *Address
	ResourcePackage *Address
	AccountPackage *Address
	IdentityPackage *Address
	ConsensusManagerPackage *Address
	AccessControllerPackage *Address
	PoolPackage *Address
	TransactionProcessorPackage *Address
	MetadataModulePackage *Address
	RoyaltyModulePackage *Address
	RoleAssignmentModulePackage *Address
	GenesisHelperPackage *Address
	FaucetPackage *Address
}

func (r *PackageAddresses) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.PackagePackage);
		FfiDestroyerAddress{}.Destroy(r.ResourcePackage);
		FfiDestroyerAddress{}.Destroy(r.AccountPackage);
		FfiDestroyerAddress{}.Destroy(r.IdentityPackage);
		FfiDestroyerAddress{}.Destroy(r.ConsensusManagerPackage);
		FfiDestroyerAddress{}.Destroy(r.AccessControllerPackage);
		FfiDestroyerAddress{}.Destroy(r.PoolPackage);
		FfiDestroyerAddress{}.Destroy(r.TransactionProcessorPackage);
		FfiDestroyerAddress{}.Destroy(r.MetadataModulePackage);
		FfiDestroyerAddress{}.Destroy(r.RoyaltyModulePackage);
		FfiDestroyerAddress{}.Destroy(r.RoleAssignmentModulePackage);
		FfiDestroyerAddress{}.Destroy(r.GenesisHelperPackage);
		FfiDestroyerAddress{}.Destroy(r.FaucetPackage);
}

type FfiConverterTypePackageAddresses struct {}

var FfiConverterTypePackageAddressesINSTANCE = FfiConverterTypePackageAddresses{}

func (c FfiConverterTypePackageAddresses) Lift(rb RustBufferI) PackageAddresses {
	return LiftFromRustBuffer[PackageAddresses](c, rb)
}

func (c FfiConverterTypePackageAddresses) Read(reader io.Reader) PackageAddresses {
	return PackageAddresses {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePackageAddresses) Lower(value PackageAddresses) RustBuffer {
	return LowerIntoRustBuffer[PackageAddresses](c, value)
}

func (c FfiConverterTypePackageAddresses) Write(writer io.Writer, value PackageAddresses) {
		FfiConverterAddressINSTANCE.Write(writer, value.PackagePackage);
		FfiConverterAddressINSTANCE.Write(writer, value.ResourcePackage);
		FfiConverterAddressINSTANCE.Write(writer, value.AccountPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.IdentityPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.ConsensusManagerPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.AccessControllerPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.PoolPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.TransactionProcessorPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.MetadataModulePackage);
		FfiConverterAddressINSTANCE.Write(writer, value.RoyaltyModulePackage);
		FfiConverterAddressINSTANCE.Write(writer, value.RoleAssignmentModulePackage);
		FfiConverterAddressINSTANCE.Write(writer, value.GenesisHelperPackage);
		FfiConverterAddressINSTANCE.Write(writer, value.FaucetPackage);
}

type FfiDestroyerTypePackageAddresses struct {}

func (_ FfiDestroyerTypePackageAddresses) Destroy(value PackageAddresses) {
	value.Destroy()
}


type PlainTextMessage struct {
	MimeType string
	Message MessageContent
}

func (r *PlainTextMessage) Destroy() {
		FfiDestroyerString{}.Destroy(r.MimeType);
		FfiDestroyerTypeMessageContent{}.Destroy(r.Message);
}

type FfiConverterTypePlainTextMessage struct {}

var FfiConverterTypePlainTextMessageINSTANCE = FfiConverterTypePlainTextMessage{}

func (c FfiConverterTypePlainTextMessage) Lift(rb RustBufferI) PlainTextMessage {
	return LiftFromRustBuffer[PlainTextMessage](c, rb)
}

func (c FfiConverterTypePlainTextMessage) Read(reader io.Reader) PlainTextMessage {
	return PlainTextMessage {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypeMessageContentINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePlainTextMessage) Lower(value PlainTextMessage) RustBuffer {
	return LowerIntoRustBuffer[PlainTextMessage](c, value)
}

func (c FfiConverterTypePlainTextMessage) Write(writer io.Writer, value PlainTextMessage) {
		FfiConverterStringINSTANCE.Write(writer, value.MimeType);
		FfiConverterTypeMessageContentINSTANCE.Write(writer, value.Message);
}

type FfiDestroyerTypePlainTextMessage struct {}

func (_ FfiDestroyerTypePlainTextMessage) Destroy(value PlainTextMessage) {
	value.Destroy()
}


type PredictedDecimal struct {
	Value *Decimal
	InstructionIndex uint64
}

func (r *PredictedDecimal) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.Value);
		FfiDestroyerUint64{}.Destroy(r.InstructionIndex);
}

type FfiConverterTypePredictedDecimal struct {}

var FfiConverterTypePredictedDecimalINSTANCE = FfiConverterTypePredictedDecimal{}

func (c FfiConverterTypePredictedDecimal) Lift(rb RustBufferI) PredictedDecimal {
	return LiftFromRustBuffer[PredictedDecimal](c, rb)
}

func (c FfiConverterTypePredictedDecimal) Read(reader io.Reader) PredictedDecimal {
	return PredictedDecimal {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePredictedDecimal) Lower(value PredictedDecimal) RustBuffer {
	return LowerIntoRustBuffer[PredictedDecimal](c, value)
}

func (c FfiConverterTypePredictedDecimal) Write(writer io.Writer, value PredictedDecimal) {
		FfiConverterDecimalINSTANCE.Write(writer, value.Value);
		FfiConverterUint64INSTANCE.Write(writer, value.InstructionIndex);
}

type FfiDestroyerTypePredictedDecimal struct {}

func (_ FfiDestroyerTypePredictedDecimal) Destroy(value PredictedDecimal) {
	value.Destroy()
}


type PredictedNonFungibleIds struct {
	Value []NonFungibleLocalId
	InstructionIndex uint64
}

func (r *PredictedNonFungibleIds) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.Value);
		FfiDestroyerUint64{}.Destroy(r.InstructionIndex);
}

type FfiConverterTypePredictedNonFungibleIds struct {}

var FfiConverterTypePredictedNonFungibleIdsINSTANCE = FfiConverterTypePredictedNonFungibleIds{}

func (c FfiConverterTypePredictedNonFungibleIds) Lift(rb RustBufferI) PredictedNonFungibleIds {
	return LiftFromRustBuffer[PredictedNonFungibleIds](c, rb)
}

func (c FfiConverterTypePredictedNonFungibleIds) Read(reader io.Reader) PredictedNonFungibleIds {
	return PredictedNonFungibleIds {
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePredictedNonFungibleIds) Lower(value PredictedNonFungibleIds) RustBuffer {
	return LowerIntoRustBuffer[PredictedNonFungibleIds](c, value)
}

func (c FfiConverterTypePredictedNonFungibleIds) Write(writer io.Writer, value PredictedNonFungibleIds) {
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.Value);
		FfiConverterUint64INSTANCE.Write(writer, value.InstructionIndex);
}

type FfiDestroyerTypePredictedNonFungibleIds struct {}

func (_ FfiDestroyerTypePredictedNonFungibleIds) Destroy(value PredictedNonFungibleIds) {
	value.Destroy()
}


type ProtocolUpdateReadinessSignalEvent struct {
	ProtocolVersionName string
}

func (r *ProtocolUpdateReadinessSignalEvent) Destroy() {
		FfiDestroyerString{}.Destroy(r.ProtocolVersionName);
}

type FfiConverterTypeProtocolUpdateReadinessSignalEvent struct {}

var FfiConverterTypeProtocolUpdateReadinessSignalEventINSTANCE = FfiConverterTypeProtocolUpdateReadinessSignalEvent{}

func (c FfiConverterTypeProtocolUpdateReadinessSignalEvent) Lift(rb RustBufferI) ProtocolUpdateReadinessSignalEvent {
	return LiftFromRustBuffer[ProtocolUpdateReadinessSignalEvent](c, rb)
}

func (c FfiConverterTypeProtocolUpdateReadinessSignalEvent) Read(reader io.Reader) ProtocolUpdateReadinessSignalEvent {
	return ProtocolUpdateReadinessSignalEvent {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeProtocolUpdateReadinessSignalEvent) Lower(value ProtocolUpdateReadinessSignalEvent) RustBuffer {
	return LowerIntoRustBuffer[ProtocolUpdateReadinessSignalEvent](c, value)
}

func (c FfiConverterTypeProtocolUpdateReadinessSignalEvent) Write(writer io.Writer, value ProtocolUpdateReadinessSignalEvent) {
		FfiConverterStringINSTANCE.Write(writer, value.ProtocolVersionName);
}

type FfiDestroyerTypeProtocolUpdateReadinessSignalEvent struct {}

func (_ FfiDestroyerTypeProtocolUpdateReadinessSignalEvent) Destroy(value ProtocolUpdateReadinessSignalEvent) {
	value.Destroy()
}


type PublicKeyFingerprint struct {
	Bytes HashableBytes
}

func (r *PublicKeyFingerprint) Destroy() {
		FfiDestroyerTypeHashableBytes{}.Destroy(r.Bytes);
}

type FfiConverterTypePublicKeyFingerprint struct {}

var FfiConverterTypePublicKeyFingerprintINSTANCE = FfiConverterTypePublicKeyFingerprint{}

func (c FfiConverterTypePublicKeyFingerprint) Lift(rb RustBufferI) PublicKeyFingerprint {
	return LiftFromRustBuffer[PublicKeyFingerprint](c, rb)
}

func (c FfiConverterTypePublicKeyFingerprint) Read(reader io.Reader) PublicKeyFingerprint {
	return PublicKeyFingerprint {
			FfiConverterTypeHashableBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePublicKeyFingerprint) Lower(value PublicKeyFingerprint) RustBuffer {
	return LowerIntoRustBuffer[PublicKeyFingerprint](c, value)
}

func (c FfiConverterTypePublicKeyFingerprint) Write(writer io.Writer, value PublicKeyFingerprint) {
		FfiConverterTypeHashableBytesINSTANCE.Write(writer, value.Bytes);
}

type FfiDestroyerTypePublicKeyFingerprint struct {}

func (_ FfiDestroyerTypePublicKeyFingerprint) Destroy(value PublicKeyFingerprint) {
	value.Destroy()
}


type RecoverEvent struct {
	Claimant *Address
	ResourceAddress *Address
	Resources ResourceSpecifier
}

func (r *RecoverEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.Claimant);
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerTypeResourceSpecifier{}.Destroy(r.Resources);
}

type FfiConverterTypeRecoverEvent struct {}

var FfiConverterTypeRecoverEventINSTANCE = FfiConverterTypeRecoverEvent{}

func (c FfiConverterTypeRecoverEvent) Lift(rb RustBufferI) RecoverEvent {
	return LiftFromRustBuffer[RecoverEvent](c, rb)
}

func (c FfiConverterTypeRecoverEvent) Read(reader io.Reader) RecoverEvent {
	return RecoverEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterTypeResourceSpecifierINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRecoverEvent) Lower(value RecoverEvent) RustBuffer {
	return LowerIntoRustBuffer[RecoverEvent](c, value)
}

func (c FfiConverterTypeRecoverEvent) Write(writer io.Writer, value RecoverEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.Claimant);
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterTypeResourceSpecifierINSTANCE.Write(writer, value.Resources);
}

type FfiDestroyerTypeRecoverEvent struct {}

func (_ FfiDestroyerTypeRecoverEvent) Destroy(value RecoverEvent) {
	value.Destroy()
}


type RecoveryProposal struct {
	RuleSet RuleSet
	TimedRecoveryDelayInMinutes *uint32
}

func (r *RecoveryProposal) Destroy() {
		FfiDestroyerTypeRuleSet{}.Destroy(r.RuleSet);
		FfiDestroyerOptionalUint32{}.Destroy(r.TimedRecoveryDelayInMinutes);
}

type FfiConverterTypeRecoveryProposal struct {}

var FfiConverterTypeRecoveryProposalINSTANCE = FfiConverterTypeRecoveryProposal{}

func (c FfiConverterTypeRecoveryProposal) Lift(rb RustBufferI) RecoveryProposal {
	return LiftFromRustBuffer[RecoveryProposal](c, rb)
}

func (c FfiConverterTypeRecoveryProposal) Read(reader io.Reader) RecoveryProposal {
	return RecoveryProposal {
			FfiConverterTypeRuleSetINSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRecoveryProposal) Lower(value RecoveryProposal) RustBuffer {
	return LowerIntoRustBuffer[RecoveryProposal](c, value)
}

func (c FfiConverterTypeRecoveryProposal) Write(writer io.Writer, value RecoveryProposal) {
		FfiConverterTypeRuleSetINSTANCE.Write(writer, value.RuleSet);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.TimedRecoveryDelayInMinutes);
}

type FfiDestroyerTypeRecoveryProposal struct {}

func (_ FfiDestroyerTypeRecoveryProposal) Destroy(value RecoveryProposal) {
	value.Destroy()
}


type RegisterValidatorEvent struct {
	PlaceholderField bool
}

func (r *RegisterValidatorEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.PlaceholderField);
}

type FfiConverterTypeRegisterValidatorEvent struct {}

var FfiConverterTypeRegisterValidatorEventINSTANCE = FfiConverterTypeRegisterValidatorEvent{}

func (c FfiConverterTypeRegisterValidatorEvent) Lift(rb RustBufferI) RegisterValidatorEvent {
	return LiftFromRustBuffer[RegisterValidatorEvent](c, rb)
}

func (c FfiConverterTypeRegisterValidatorEvent) Read(reader io.Reader) RegisterValidatorEvent {
	return RegisterValidatorEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRegisterValidatorEvent) Lower(value RegisterValidatorEvent) RustBuffer {
	return LowerIntoRustBuffer[RegisterValidatorEvent](c, value)
}

func (c FfiConverterTypeRegisterValidatorEvent) Write(writer io.Writer, value RegisterValidatorEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.PlaceholderField);
}

type FfiDestroyerTypeRegisterValidatorEvent struct {}

func (_ FfiDestroyerTypeRegisterValidatorEvent) Destroy(value RegisterValidatorEvent) {
	value.Destroy()
}


type RemoveMetadataEvent struct {
	Key string
}

func (r *RemoveMetadataEvent) Destroy() {
		FfiDestroyerString{}.Destroy(r.Key);
}

type FfiConverterTypeRemoveMetadataEvent struct {}

var FfiConverterTypeRemoveMetadataEventINSTANCE = FfiConverterTypeRemoveMetadataEvent{}

func (c FfiConverterTypeRemoveMetadataEvent) Lift(rb RustBufferI) RemoveMetadataEvent {
	return LiftFromRustBuffer[RemoveMetadataEvent](c, rb)
}

func (c FfiConverterTypeRemoveMetadataEvent) Read(reader io.Reader) RemoveMetadataEvent {
	return RemoveMetadataEvent {
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRemoveMetadataEvent) Lower(value RemoveMetadataEvent) RustBuffer {
	return LowerIntoRustBuffer[RemoveMetadataEvent](c, value)
}

func (c FfiConverterTypeRemoveMetadataEvent) Write(writer io.Writer, value RemoveMetadataEvent) {
		FfiConverterStringINSTANCE.Write(writer, value.Key);
}

type FfiDestroyerTypeRemoveMetadataEvent struct {}

func (_ FfiDestroyerTypeRemoveMetadataEvent) Destroy(value RemoveMetadataEvent) {
	value.Destroy()
}


type ResourceAddresses struct {
	Xrd *Address
	Secp256k1SignatureVirtualBadge *Address
	Ed25519SignatureVirtualBadge *Address
	PackageOfDirectCallerVirtualBadge *Address
	GlobalCallerVirtualBadge *Address
	SystemTransactionBadge *Address
	PackageOwnerBadge *Address
	ValidatorOwnerBadge *Address
	AccountOwnerBadge *Address
	IdentityOwnerBadge *Address
}

func (r *ResourceAddresses) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.Xrd);
		FfiDestroyerAddress{}.Destroy(r.Secp256k1SignatureVirtualBadge);
		FfiDestroyerAddress{}.Destroy(r.Ed25519SignatureVirtualBadge);
		FfiDestroyerAddress{}.Destroy(r.PackageOfDirectCallerVirtualBadge);
		FfiDestroyerAddress{}.Destroy(r.GlobalCallerVirtualBadge);
		FfiDestroyerAddress{}.Destroy(r.SystemTransactionBadge);
		FfiDestroyerAddress{}.Destroy(r.PackageOwnerBadge);
		FfiDestroyerAddress{}.Destroy(r.ValidatorOwnerBadge);
		FfiDestroyerAddress{}.Destroy(r.AccountOwnerBadge);
		FfiDestroyerAddress{}.Destroy(r.IdentityOwnerBadge);
}

type FfiConverterTypeResourceAddresses struct {}

var FfiConverterTypeResourceAddressesINSTANCE = FfiConverterTypeResourceAddresses{}

func (c FfiConverterTypeResourceAddresses) Lift(rb RustBufferI) ResourceAddresses {
	return LiftFromRustBuffer[ResourceAddresses](c, rb)
}

func (c FfiConverterTypeResourceAddresses) Read(reader io.Reader) ResourceAddresses {
	return ResourceAddresses {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeResourceAddresses) Lower(value ResourceAddresses) RustBuffer {
	return LowerIntoRustBuffer[ResourceAddresses](c, value)
}

func (c FfiConverterTypeResourceAddresses) Write(writer io.Writer, value ResourceAddresses) {
		FfiConverterAddressINSTANCE.Write(writer, value.Xrd);
		FfiConverterAddressINSTANCE.Write(writer, value.Secp256k1SignatureVirtualBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.Ed25519SignatureVirtualBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.PackageOfDirectCallerVirtualBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.GlobalCallerVirtualBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.SystemTransactionBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.PackageOwnerBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.ValidatorOwnerBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.AccountOwnerBadge);
		FfiConverterAddressINSTANCE.Write(writer, value.IdentityOwnerBadge);
}

type FfiDestroyerTypeResourceAddresses struct {}

func (_ FfiDestroyerTypeResourceAddresses) Destroy(value ResourceAddresses) {
	value.Destroy()
}


type ResourceManagerRole struct {
	Role **AccessRule
	RoleUpdater **AccessRule
}

func (r *ResourceManagerRole) Destroy() {
		FfiDestroyerOptionalAccessRule{}.Destroy(r.Role);
		FfiDestroyerOptionalAccessRule{}.Destroy(r.RoleUpdater);
}

type FfiConverterTypeResourceManagerRole struct {}

var FfiConverterTypeResourceManagerRoleINSTANCE = FfiConverterTypeResourceManagerRole{}

func (c FfiConverterTypeResourceManagerRole) Lift(rb RustBufferI) ResourceManagerRole {
	return LiftFromRustBuffer[ResourceManagerRole](c, rb)
}

func (c FfiConverterTypeResourceManagerRole) Read(reader io.Reader) ResourceManagerRole {
	return ResourceManagerRole {
			FfiConverterOptionalAccessRuleINSTANCE.Read(reader),
			FfiConverterOptionalAccessRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeResourceManagerRole) Lower(value ResourceManagerRole) RustBuffer {
	return LowerIntoRustBuffer[ResourceManagerRole](c, value)
}

func (c FfiConverterTypeResourceManagerRole) Write(writer io.Writer, value ResourceManagerRole) {
		FfiConverterOptionalAccessRuleINSTANCE.Write(writer, value.Role);
		FfiConverterOptionalAccessRuleINSTANCE.Write(writer, value.RoleUpdater);
}

type FfiDestroyerTypeResourceManagerRole struct {}

func (_ FfiDestroyerTypeResourceManagerRole) Destroy(value ResourceManagerRole) {
	value.Destroy()
}


type RoundChangeEvent struct {
	Round uint64
}

func (r *RoundChangeEvent) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.Round);
}

type FfiConverterTypeRoundChangeEvent struct {}

var FfiConverterTypeRoundChangeEventINSTANCE = FfiConverterTypeRoundChangeEvent{}

func (c FfiConverterTypeRoundChangeEvent) Lift(rb RustBufferI) RoundChangeEvent {
	return LiftFromRustBuffer[RoundChangeEvent](c, rb)
}

func (c FfiConverterTypeRoundChangeEvent) Read(reader io.Reader) RoundChangeEvent {
	return RoundChangeEvent {
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRoundChangeEvent) Lower(value RoundChangeEvent) RustBuffer {
	return LowerIntoRustBuffer[RoundChangeEvent](c, value)
}

func (c FfiConverterTypeRoundChangeEvent) Write(writer io.Writer, value RoundChangeEvent) {
		FfiConverterUint64INSTANCE.Write(writer, value.Round);
}

type FfiDestroyerTypeRoundChangeEvent struct {}

func (_ FfiDestroyerTypeRoundChangeEvent) Destroy(value RoundChangeEvent) {
	value.Destroy()
}


type RuleSet struct {
	PrimaryRole *AccessRule
	RecoveryRole *AccessRule
	ConfirmationRole *AccessRule
}

func (r *RuleSet) Destroy() {
		FfiDestroyerAccessRule{}.Destroy(r.PrimaryRole);
		FfiDestroyerAccessRule{}.Destroy(r.RecoveryRole);
		FfiDestroyerAccessRule{}.Destroy(r.ConfirmationRole);
}

type FfiConverterTypeRuleSet struct {}

var FfiConverterTypeRuleSetINSTANCE = FfiConverterTypeRuleSet{}

func (c FfiConverterTypeRuleSet) Lift(rb RustBufferI) RuleSet {
	return LiftFromRustBuffer[RuleSet](c, rb)
}

func (c FfiConverterTypeRuleSet) Read(reader io.Reader) RuleSet {
	return RuleSet {
			FfiConverterAccessRuleINSTANCE.Read(reader),
			FfiConverterAccessRuleINSTANCE.Read(reader),
			FfiConverterAccessRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRuleSet) Lower(value RuleSet) RustBuffer {
	return LowerIntoRustBuffer[RuleSet](c, value)
}

func (c FfiConverterTypeRuleSet) Write(writer io.Writer, value RuleSet) {
		FfiConverterAccessRuleINSTANCE.Write(writer, value.PrimaryRole);
		FfiConverterAccessRuleINSTANCE.Write(writer, value.RecoveryRole);
		FfiConverterAccessRuleINSTANCE.Write(writer, value.ConfirmationRole);
}

type FfiDestroyerTypeRuleSet struct {}

func (_ FfiDestroyerTypeRuleSet) Destroy(value RuleSet) {
	value.Destroy()
}


type RuleSetUpdateEvent struct {
	Proposer Proposer
	Proposal RecoveryProposal
}

func (r *RuleSetUpdateEvent) Destroy() {
		FfiDestroyerTypeProposer{}.Destroy(r.Proposer);
		FfiDestroyerTypeRecoveryProposal{}.Destroy(r.Proposal);
}

type FfiConverterTypeRuleSetUpdateEvent struct {}

var FfiConverterTypeRuleSetUpdateEventINSTANCE = FfiConverterTypeRuleSetUpdateEvent{}

func (c FfiConverterTypeRuleSetUpdateEvent) Lift(rb RustBufferI) RuleSetUpdateEvent {
	return LiftFromRustBuffer[RuleSetUpdateEvent](c, rb)
}

func (c FfiConverterTypeRuleSetUpdateEvent) Read(reader io.Reader) RuleSetUpdateEvent {
	return RuleSetUpdateEvent {
			FfiConverterTypeProposerINSTANCE.Read(reader),
			FfiConverterTypeRecoveryProposalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRuleSetUpdateEvent) Lower(value RuleSetUpdateEvent) RustBuffer {
	return LowerIntoRustBuffer[RuleSetUpdateEvent](c, value)
}

func (c FfiConverterTypeRuleSetUpdateEvent) Write(writer io.Writer, value RuleSetUpdateEvent) {
		FfiConverterTypeProposerINSTANCE.Write(writer, value.Proposer);
		FfiConverterTypeRecoveryProposalINSTANCE.Write(writer, value.Proposal);
}

type FfiDestroyerTypeRuleSetUpdateEvent struct {}

func (_ FfiDestroyerTypeRuleSetUpdateEvent) Destroy(value RuleSetUpdateEvent) {
	value.Destroy()
}


type Schema struct {
	LocalTypeId LocalTypeId
	Schema []byte
}

func (r *Schema) Destroy() {
		FfiDestroyerTypeLocalTypeId{}.Destroy(r.LocalTypeId);
		FfiDestroyerBytes{}.Destroy(r.Schema);
}

type FfiConverterTypeSchema struct {}

var FfiConverterTypeSchemaINSTANCE = FfiConverterTypeSchema{}

func (c FfiConverterTypeSchema) Lift(rb RustBufferI) Schema {
	return LiftFromRustBuffer[Schema](c, rb)
}

func (c FfiConverterTypeSchema) Read(reader io.Reader) Schema {
	return Schema {
			FfiConverterTypeLocalTypeIdINSTANCE.Read(reader),
			FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSchema) Lower(value Schema) RustBuffer {
	return LowerIntoRustBuffer[Schema](c, value)
}

func (c FfiConverterTypeSchema) Write(writer io.Writer, value Schema) {
		FfiConverterTypeLocalTypeIdINSTANCE.Write(writer, value.LocalTypeId);
		FfiConverterBytesINSTANCE.Write(writer, value.Schema);
}

type FfiDestroyerTypeSchema struct {}

func (_ FfiDestroyerTypeSchema) Destroy(value Schema) {
	value.Destroy()
}


type Secp256k1PublicKey struct {
	Value []byte
}

func (r *Secp256k1PublicKey) Destroy() {
		FfiDestroyerBytes{}.Destroy(r.Value);
}

type FfiConverterTypeSecp256k1PublicKey struct {}

var FfiConverterTypeSecp256k1PublicKeyINSTANCE = FfiConverterTypeSecp256k1PublicKey{}

func (c FfiConverterTypeSecp256k1PublicKey) Lift(rb RustBufferI) Secp256k1PublicKey {
	return LiftFromRustBuffer[Secp256k1PublicKey](c, rb)
}

func (c FfiConverterTypeSecp256k1PublicKey) Read(reader io.Reader) Secp256k1PublicKey {
	return Secp256k1PublicKey {
			FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSecp256k1PublicKey) Lower(value Secp256k1PublicKey) RustBuffer {
	return LowerIntoRustBuffer[Secp256k1PublicKey](c, value)
}

func (c FfiConverterTypeSecp256k1PublicKey) Write(writer io.Writer, value Secp256k1PublicKey) {
		FfiConverterBytesINSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeSecp256k1PublicKey struct {}

func (_ FfiDestroyerTypeSecp256k1PublicKey) Destroy(value Secp256k1PublicKey) {
	value.Destroy()
}


type SecurityStructureRole struct {
	SuperAdminFactors []PublicKey
	ThresholdFactors []PublicKey
	Threshold uint8
}

func (r *SecurityStructureRole) Destroy() {
		FfiDestroyerSequenceTypePublicKey{}.Destroy(r.SuperAdminFactors);
		FfiDestroyerSequenceTypePublicKey{}.Destroy(r.ThresholdFactors);
		FfiDestroyerUint8{}.Destroy(r.Threshold);
}

type FfiConverterTypeSecurityStructureRole struct {}

var FfiConverterTypeSecurityStructureRoleINSTANCE = FfiConverterTypeSecurityStructureRole{}

func (c FfiConverterTypeSecurityStructureRole) Lift(rb RustBufferI) SecurityStructureRole {
	return LiftFromRustBuffer[SecurityStructureRole](c, rb)
}

func (c FfiConverterTypeSecurityStructureRole) Read(reader io.Reader) SecurityStructureRole {
	return SecurityStructureRole {
			FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
			FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint8INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSecurityStructureRole) Lower(value SecurityStructureRole) RustBuffer {
	return LowerIntoRustBuffer[SecurityStructureRole](c, value)
}

func (c FfiConverterTypeSecurityStructureRole) Write(writer io.Writer, value SecurityStructureRole) {
		FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, value.SuperAdminFactors);
		FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, value.ThresholdFactors);
		FfiConverterUint8INSTANCE.Write(writer, value.Threshold);
}

type FfiDestroyerTypeSecurityStructureRole struct {}

func (_ FfiDestroyerTypeSecurityStructureRole) Destroy(value SecurityStructureRole) {
	value.Destroy()
}


type SetAndLockRoleEvent struct {
	RoleKey string
	Rule *AccessRule
}

func (r *SetAndLockRoleEvent) Destroy() {
		FfiDestroyerString{}.Destroy(r.RoleKey);
		FfiDestroyerAccessRule{}.Destroy(r.Rule);
}

type FfiConverterTypeSetAndLockRoleEvent struct {}

var FfiConverterTypeSetAndLockRoleEventINSTANCE = FfiConverterTypeSetAndLockRoleEvent{}

func (c FfiConverterTypeSetAndLockRoleEvent) Lift(rb RustBufferI) SetAndLockRoleEvent {
	return LiftFromRustBuffer[SetAndLockRoleEvent](c, rb)
}

func (c FfiConverterTypeSetAndLockRoleEvent) Read(reader io.Reader) SetAndLockRoleEvent {
	return SetAndLockRoleEvent {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterAccessRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSetAndLockRoleEvent) Lower(value SetAndLockRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[SetAndLockRoleEvent](c, value)
}

func (c FfiConverterTypeSetAndLockRoleEvent) Write(writer io.Writer, value SetAndLockRoleEvent) {
		FfiConverterStringINSTANCE.Write(writer, value.RoleKey);
		FfiConverterAccessRuleINSTANCE.Write(writer, value.Rule);
}

type FfiDestroyerTypeSetAndLockRoleEvent struct {}

func (_ FfiDestroyerTypeSetAndLockRoleEvent) Destroy(value SetAndLockRoleEvent) {
	value.Destroy()
}


type SetMetadataEvent struct {
	Key string
	Value MetadataValue
}

func (r *SetMetadataEvent) Destroy() {
		FfiDestroyerString{}.Destroy(r.Key);
		FfiDestroyerTypeMetadataValue{}.Destroy(r.Value);
}

type FfiConverterTypeSetMetadataEvent struct {}

var FfiConverterTypeSetMetadataEventINSTANCE = FfiConverterTypeSetMetadataEvent{}

func (c FfiConverterTypeSetMetadataEvent) Lift(rb RustBufferI) SetMetadataEvent {
	return LiftFromRustBuffer[SetMetadataEvent](c, rb)
}

func (c FfiConverterTypeSetMetadataEvent) Read(reader io.Reader) SetMetadataEvent {
	return SetMetadataEvent {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypeMetadataValueINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSetMetadataEvent) Lower(value SetMetadataEvent) RustBuffer {
	return LowerIntoRustBuffer[SetMetadataEvent](c, value)
}

func (c FfiConverterTypeSetMetadataEvent) Write(writer io.Writer, value SetMetadataEvent) {
		FfiConverterStringINSTANCE.Write(writer, value.Key);
		FfiConverterTypeMetadataValueINSTANCE.Write(writer, value.Value);
}

type FfiDestroyerTypeSetMetadataEvent struct {}

func (_ FfiDestroyerTypeSetMetadataEvent) Destroy(value SetMetadataEvent) {
	value.Destroy()
}


type SetOwnerRoleEvent struct {
	Rule *AccessRule
}

func (r *SetOwnerRoleEvent) Destroy() {
		FfiDestroyerAccessRule{}.Destroy(r.Rule);
}

type FfiConverterTypeSetOwnerRoleEvent struct {}

var FfiConverterTypeSetOwnerRoleEventINSTANCE = FfiConverterTypeSetOwnerRoleEvent{}

func (c FfiConverterTypeSetOwnerRoleEvent) Lift(rb RustBufferI) SetOwnerRoleEvent {
	return LiftFromRustBuffer[SetOwnerRoleEvent](c, rb)
}

func (c FfiConverterTypeSetOwnerRoleEvent) Read(reader io.Reader) SetOwnerRoleEvent {
	return SetOwnerRoleEvent {
			FfiConverterAccessRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSetOwnerRoleEvent) Lower(value SetOwnerRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[SetOwnerRoleEvent](c, value)
}

func (c FfiConverterTypeSetOwnerRoleEvent) Write(writer io.Writer, value SetOwnerRoleEvent) {
		FfiConverterAccessRuleINSTANCE.Write(writer, value.Rule);
}

type FfiDestroyerTypeSetOwnerRoleEvent struct {}

func (_ FfiDestroyerTypeSetOwnerRoleEvent) Destroy(value SetOwnerRoleEvent) {
	value.Destroy()
}


type SetRoleEvent struct {
	RoleKey string
	Rule *AccessRule
}

func (r *SetRoleEvent) Destroy() {
		FfiDestroyerString{}.Destroy(r.RoleKey);
		FfiDestroyerAccessRule{}.Destroy(r.Rule);
}

type FfiConverterTypeSetRoleEvent struct {}

var FfiConverterTypeSetRoleEventINSTANCE = FfiConverterTypeSetRoleEvent{}

func (c FfiConverterTypeSetRoleEvent) Lift(rb RustBufferI) SetRoleEvent {
	return LiftFromRustBuffer[SetRoleEvent](c, rb)
}

func (c FfiConverterTypeSetRoleEvent) Read(reader io.Reader) SetRoleEvent {
	return SetRoleEvent {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterAccessRuleINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSetRoleEvent) Lower(value SetRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[SetRoleEvent](c, value)
}

func (c FfiConverterTypeSetRoleEvent) Write(writer io.Writer, value SetRoleEvent) {
		FfiConverterStringINSTANCE.Write(writer, value.RoleKey);
		FfiConverterAccessRuleINSTANCE.Write(writer, value.Rule);
}

type FfiDestroyerTypeSetRoleEvent struct {}

func (_ FfiDestroyerTypeSetRoleEvent) Destroy(value SetRoleEvent) {
	value.Destroy()
}


type StakeEvent struct {
	XrdStaked *Decimal
}

func (r *StakeEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.XrdStaked);
}

type FfiConverterTypeStakeEvent struct {}

var FfiConverterTypeStakeEventINSTANCE = FfiConverterTypeStakeEvent{}

func (c FfiConverterTypeStakeEvent) Lift(rb RustBufferI) StakeEvent {
	return LiftFromRustBuffer[StakeEvent](c, rb)
}

func (c FfiConverterTypeStakeEvent) Read(reader io.Reader) StakeEvent {
	return StakeEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeStakeEvent) Lower(value StakeEvent) RustBuffer {
	return LowerIntoRustBuffer[StakeEvent](c, value)
}

func (c FfiConverterTypeStakeEvent) Write(writer io.Writer, value StakeEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.XrdStaked);
}

type FfiDestroyerTypeStakeEvent struct {}

func (_ FfiDestroyerTypeStakeEvent) Destroy(value StakeEvent) {
	value.Destroy()
}


type StopTimedRecoveryEvent struct {
	PlaceholderField bool
}

func (r *StopTimedRecoveryEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.PlaceholderField);
}

type FfiConverterTypeStopTimedRecoveryEvent struct {}

var FfiConverterTypeStopTimedRecoveryEventINSTANCE = FfiConverterTypeStopTimedRecoveryEvent{}

func (c FfiConverterTypeStopTimedRecoveryEvent) Lift(rb RustBufferI) StopTimedRecoveryEvent {
	return LiftFromRustBuffer[StopTimedRecoveryEvent](c, rb)
}

func (c FfiConverterTypeStopTimedRecoveryEvent) Read(reader io.Reader) StopTimedRecoveryEvent {
	return StopTimedRecoveryEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeStopTimedRecoveryEvent) Lower(value StopTimedRecoveryEvent) RustBuffer {
	return LowerIntoRustBuffer[StopTimedRecoveryEvent](c, value)
}

func (c FfiConverterTypeStopTimedRecoveryEvent) Write(writer io.Writer, value StopTimedRecoveryEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.PlaceholderField);
}

type FfiDestroyerTypeStopTimedRecoveryEvent struct {}

func (_ FfiDestroyerTypeStopTimedRecoveryEvent) Destroy(value StopTimedRecoveryEvent) {
	value.Destroy()
}


type StoreEvent struct {
	Claimant *Address
	ResourceAddress *Address
	Resources ResourceSpecifier
}

func (r *StoreEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.Claimant);
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerTypeResourceSpecifier{}.Destroy(r.Resources);
}

type FfiConverterTypeStoreEvent struct {}

var FfiConverterTypeStoreEventINSTANCE = FfiConverterTypeStoreEvent{}

func (c FfiConverterTypeStoreEvent) Lift(rb RustBufferI) StoreEvent {
	return LiftFromRustBuffer[StoreEvent](c, rb)
}

func (c FfiConverterTypeStoreEvent) Read(reader io.Reader) StoreEvent {
	return StoreEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterTypeResourceSpecifierINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeStoreEvent) Lower(value StoreEvent) RustBuffer {
	return LowerIntoRustBuffer[StoreEvent](c, value)
}

func (c FfiConverterTypeStoreEvent) Write(writer io.Writer, value StoreEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.Claimant);
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterTypeResourceSpecifierINSTANCE.Write(writer, value.Resources);
}

type FfiDestroyerTypeStoreEvent struct {}

func (_ FfiDestroyerTypeStoreEvent) Destroy(value StoreEvent) {
	value.Destroy()
}


type TrackedPoolContribution struct {
	PoolAddress *Address
	ContributedResources map[string]*Decimal
	PoolUnitsResourceAddress *Address
	PoolUnitsAmount *Decimal
}

func (r *TrackedPoolContribution) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.PoolAddress);
		FfiDestroyerMapStringDecimal{}.Destroy(r.ContributedResources);
		FfiDestroyerAddress{}.Destroy(r.PoolUnitsResourceAddress);
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitsAmount);
}

type FfiConverterTypeTrackedPoolContribution struct {}

var FfiConverterTypeTrackedPoolContributionINSTANCE = FfiConverterTypeTrackedPoolContribution{}

func (c FfiConverterTypeTrackedPoolContribution) Lift(rb RustBufferI) TrackedPoolContribution {
	return LiftFromRustBuffer[TrackedPoolContribution](c, rb)
}

func (c FfiConverterTypeTrackedPoolContribution) Read(reader io.Reader) TrackedPoolContribution {
	return TrackedPoolContribution {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterMapStringDecimalINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTrackedPoolContribution) Lower(value TrackedPoolContribution) RustBuffer {
	return LowerIntoRustBuffer[TrackedPoolContribution](c, value)
}

func (c FfiConverterTypeTrackedPoolContribution) Write(writer io.Writer, value TrackedPoolContribution) {
		FfiConverterAddressINSTANCE.Write(writer, value.PoolAddress);
		FfiConverterMapStringDecimalINSTANCE.Write(writer, value.ContributedResources);
		FfiConverterAddressINSTANCE.Write(writer, value.PoolUnitsResourceAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitsAmount);
}

type FfiDestroyerTypeTrackedPoolContribution struct {}

func (_ FfiDestroyerTypeTrackedPoolContribution) Destroy(value TrackedPoolContribution) {
	value.Destroy()
}


type TrackedPoolRedemption struct {
	PoolAddress *Address
	PoolUnitsResourceAddress *Address
	PoolUnitsAmount *Decimal
	RedeemedResources map[string]*Decimal
}

func (r *TrackedPoolRedemption) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.PoolAddress);
		FfiDestroyerAddress{}.Destroy(r.PoolUnitsResourceAddress);
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitsAmount);
		FfiDestroyerMapStringDecimal{}.Destroy(r.RedeemedResources);
}

type FfiConverterTypeTrackedPoolRedemption struct {}

var FfiConverterTypeTrackedPoolRedemptionINSTANCE = FfiConverterTypeTrackedPoolRedemption{}

func (c FfiConverterTypeTrackedPoolRedemption) Lift(rb RustBufferI) TrackedPoolRedemption {
	return LiftFromRustBuffer[TrackedPoolRedemption](c, rb)
}

func (c FfiConverterTypeTrackedPoolRedemption) Read(reader io.Reader) TrackedPoolRedemption {
	return TrackedPoolRedemption {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterMapStringDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTrackedPoolRedemption) Lower(value TrackedPoolRedemption) RustBuffer {
	return LowerIntoRustBuffer[TrackedPoolRedemption](c, value)
}

func (c FfiConverterTypeTrackedPoolRedemption) Write(writer io.Writer, value TrackedPoolRedemption) {
		FfiConverterAddressINSTANCE.Write(writer, value.PoolAddress);
		FfiConverterAddressINSTANCE.Write(writer, value.PoolUnitsResourceAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitsAmount);
		FfiConverterMapStringDecimalINSTANCE.Write(writer, value.RedeemedResources);
}

type FfiDestroyerTypeTrackedPoolRedemption struct {}

func (_ FfiDestroyerTypeTrackedPoolRedemption) Destroy(value TrackedPoolRedemption) {
	value.Destroy()
}


type TrackedValidatorClaim struct {
	ValidatorAddress *Address
	ClaimNftAddress *Address
	ClaimNftIds []NonFungibleLocalId
	XrdAmount *Decimal
}

func (r *TrackedValidatorClaim) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ValidatorAddress);
		FfiDestroyerAddress{}.Destroy(r.ClaimNftAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.ClaimNftIds);
		FfiDestroyerDecimal{}.Destroy(r.XrdAmount);
}

type FfiConverterTypeTrackedValidatorClaim struct {}

var FfiConverterTypeTrackedValidatorClaimINSTANCE = FfiConverterTypeTrackedValidatorClaim{}

func (c FfiConverterTypeTrackedValidatorClaim) Lift(rb RustBufferI) TrackedValidatorClaim {
	return LiftFromRustBuffer[TrackedValidatorClaim](c, rb)
}

func (c FfiConverterTypeTrackedValidatorClaim) Read(reader io.Reader) TrackedValidatorClaim {
	return TrackedValidatorClaim {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTrackedValidatorClaim) Lower(value TrackedValidatorClaim) RustBuffer {
	return LowerIntoRustBuffer[TrackedValidatorClaim](c, value)
}

func (c FfiConverterTypeTrackedValidatorClaim) Write(writer io.Writer, value TrackedValidatorClaim) {
		FfiConverterAddressINSTANCE.Write(writer, value.ValidatorAddress);
		FfiConverterAddressINSTANCE.Write(writer, value.ClaimNftAddress);
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.ClaimNftIds);
		FfiConverterDecimalINSTANCE.Write(writer, value.XrdAmount);
}

type FfiDestroyerTypeTrackedValidatorClaim struct {}

func (_ FfiDestroyerTypeTrackedValidatorClaim) Destroy(value TrackedValidatorClaim) {
	value.Destroy()
}


type TrackedValidatorStake struct {
	ValidatorAddress *Address
	XrdAmount *Decimal
	LiquidStakeUnitAddress *Address
	LiquidStakeUnitAmount *Decimal
}

func (r *TrackedValidatorStake) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ValidatorAddress);
		FfiDestroyerDecimal{}.Destroy(r.XrdAmount);
		FfiDestroyerAddress{}.Destroy(r.LiquidStakeUnitAddress);
		FfiDestroyerDecimal{}.Destroy(r.LiquidStakeUnitAmount);
}

type FfiConverterTypeTrackedValidatorStake struct {}

var FfiConverterTypeTrackedValidatorStakeINSTANCE = FfiConverterTypeTrackedValidatorStake{}

func (c FfiConverterTypeTrackedValidatorStake) Lift(rb RustBufferI) TrackedValidatorStake {
	return LiftFromRustBuffer[TrackedValidatorStake](c, rb)
}

func (c FfiConverterTypeTrackedValidatorStake) Read(reader io.Reader) TrackedValidatorStake {
	return TrackedValidatorStake {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTrackedValidatorStake) Lower(value TrackedValidatorStake) RustBuffer {
	return LowerIntoRustBuffer[TrackedValidatorStake](c, value)
}

func (c FfiConverterTypeTrackedValidatorStake) Write(writer io.Writer, value TrackedValidatorStake) {
		FfiConverterAddressINSTANCE.Write(writer, value.ValidatorAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.XrdAmount);
		FfiConverterAddressINSTANCE.Write(writer, value.LiquidStakeUnitAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.LiquidStakeUnitAmount);
}

type FfiDestroyerTypeTrackedValidatorStake struct {}

func (_ FfiDestroyerTypeTrackedValidatorStake) Destroy(value TrackedValidatorStake) {
	value.Destroy()
}


type TrackedValidatorUnstake struct {
	ValidatorAddress *Address
	LiquidStakeUnitAddress *Address
	LiquidStakeUnitAmount *Decimal
	ClaimNftAddress *Address
	ClaimNftIds []NonFungibleLocalId
}

func (r *TrackedValidatorUnstake) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ValidatorAddress);
		FfiDestroyerAddress{}.Destroy(r.LiquidStakeUnitAddress);
		FfiDestroyerDecimal{}.Destroy(r.LiquidStakeUnitAmount);
		FfiDestroyerAddress{}.Destroy(r.ClaimNftAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(r.ClaimNftIds);
}

type FfiConverterTypeTrackedValidatorUnstake struct {}

var FfiConverterTypeTrackedValidatorUnstakeINSTANCE = FfiConverterTypeTrackedValidatorUnstake{}

func (c FfiConverterTypeTrackedValidatorUnstake) Lift(rb RustBufferI) TrackedValidatorUnstake {
	return LiftFromRustBuffer[TrackedValidatorUnstake](c, rb)
}

func (c FfiConverterTypeTrackedValidatorUnstake) Read(reader io.Reader) TrackedValidatorUnstake {
	return TrackedValidatorUnstake {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTrackedValidatorUnstake) Lower(value TrackedValidatorUnstake) RustBuffer {
	return LowerIntoRustBuffer[TrackedValidatorUnstake](c, value)
}

func (c FfiConverterTypeTrackedValidatorUnstake) Write(writer io.Writer, value TrackedValidatorUnstake) {
		FfiConverterAddressINSTANCE.Write(writer, value.ValidatorAddress);
		FfiConverterAddressINSTANCE.Write(writer, value.LiquidStakeUnitAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.LiquidStakeUnitAmount);
		FfiConverterAddressINSTANCE.Write(writer, value.ClaimNftAddress);
		FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, value.ClaimNftIds);
}

type FfiDestroyerTypeTrackedValidatorUnstake struct {}

func (_ FfiDestroyerTypeTrackedValidatorUnstake) Destroy(value TrackedValidatorUnstake) {
	value.Destroy()
}


type TransactionHeader struct {
	NetworkId uint8
	StartEpochInclusive uint64
	EndEpochExclusive uint64
	Nonce uint32
	NotaryPublicKey PublicKey
	NotaryIsSignatory bool
	TipPercentage uint16
}

func (r *TransactionHeader) Destroy() {
		FfiDestroyerUint8{}.Destroy(r.NetworkId);
		FfiDestroyerUint64{}.Destroy(r.StartEpochInclusive);
		FfiDestroyerUint64{}.Destroy(r.EndEpochExclusive);
		FfiDestroyerUint32{}.Destroy(r.Nonce);
		FfiDestroyerTypePublicKey{}.Destroy(r.NotaryPublicKey);
		FfiDestroyerBool{}.Destroy(r.NotaryIsSignatory);
		FfiDestroyerUint16{}.Destroy(r.TipPercentage);
}

type FfiConverterTypeTransactionHeader struct {}

var FfiConverterTypeTransactionHeaderINSTANCE = FfiConverterTypeTransactionHeader{}

func (c FfiConverterTypeTransactionHeader) Lift(rb RustBufferI) TransactionHeader {
	return LiftFromRustBuffer[TransactionHeader](c, rb)
}

func (c FfiConverterTypeTransactionHeader) Read(reader io.Reader) TransactionHeader {
	return TransactionHeader {
			FfiConverterUint8INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTransactionHeader) Lower(value TransactionHeader) RustBuffer {
	return LowerIntoRustBuffer[TransactionHeader](c, value)
}

func (c FfiConverterTypeTransactionHeader) Write(writer io.Writer, value TransactionHeader) {
		FfiConverterUint8INSTANCE.Write(writer, value.NetworkId);
		FfiConverterUint64INSTANCE.Write(writer, value.StartEpochInclusive);
		FfiConverterUint64INSTANCE.Write(writer, value.EndEpochExclusive);
		FfiConverterUint32INSTANCE.Write(writer, value.Nonce);
		FfiConverterTypePublicKeyINSTANCE.Write(writer, value.NotaryPublicKey);
		FfiConverterBoolINSTANCE.Write(writer, value.NotaryIsSignatory);
		FfiConverterUint16INSTANCE.Write(writer, value.TipPercentage);
}

type FfiDestroyerTypeTransactionHeader struct {}

func (_ FfiDestroyerTypeTransactionHeader) Destroy(value TransactionHeader) {
	value.Destroy()
}


type TransactionManifestModifications struct {
	AddAccessControllerProofs []*Address
	AddLockFee *LockFeeModification
	AddAssertions []IndexedAssertion
}

func (r *TransactionManifestModifications) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(r.AddAccessControllerProofs);
		FfiDestroyerOptionalTypeLockFeeModification{}.Destroy(r.AddLockFee);
		FfiDestroyerSequenceTypeIndexedAssertion{}.Destroy(r.AddAssertions);
}

type FfiConverterTypeTransactionManifestModifications struct {}

var FfiConverterTypeTransactionManifestModificationsINSTANCE = FfiConverterTypeTransactionManifestModifications{}

func (c FfiConverterTypeTransactionManifestModifications) Lift(rb RustBufferI) TransactionManifestModifications {
	return LiftFromRustBuffer[TransactionManifestModifications](c, rb)
}

func (c FfiConverterTypeTransactionManifestModifications) Read(reader io.Reader) TransactionManifestModifications {
	return TransactionManifestModifications {
			FfiConverterSequenceAddressINSTANCE.Read(reader),
			FfiConverterOptionalTypeLockFeeModificationINSTANCE.Read(reader),
			FfiConverterSequenceTypeIndexedAssertionINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTransactionManifestModifications) Lower(value TransactionManifestModifications) RustBuffer {
	return LowerIntoRustBuffer[TransactionManifestModifications](c, value)
}

func (c FfiConverterTypeTransactionManifestModifications) Write(writer io.Writer, value TransactionManifestModifications) {
		FfiConverterSequenceAddressINSTANCE.Write(writer, value.AddAccessControllerProofs);
		FfiConverterOptionalTypeLockFeeModificationINSTANCE.Write(writer, value.AddLockFee);
		FfiConverterSequenceTypeIndexedAssertionINSTANCE.Write(writer, value.AddAssertions);
}

type FfiDestroyerTypeTransactionManifestModifications struct {}

func (_ FfiDestroyerTypeTransactionManifestModifications) Destroy(value TransactionManifestModifications) {
	value.Destroy()
}


type TwoResourcePoolContributionEvent struct {
	ContributedResources map[string]*Decimal
	PoolUnitsMinted *Decimal
}

func (r *TwoResourcePoolContributionEvent) Destroy() {
		FfiDestroyerMapStringDecimal{}.Destroy(r.ContributedResources);
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitsMinted);
}

type FfiConverterTypeTwoResourcePoolContributionEvent struct {}

var FfiConverterTypeTwoResourcePoolContributionEventINSTANCE = FfiConverterTypeTwoResourcePoolContributionEvent{}

func (c FfiConverterTypeTwoResourcePoolContributionEvent) Lift(rb RustBufferI) TwoResourcePoolContributionEvent {
	return LiftFromRustBuffer[TwoResourcePoolContributionEvent](c, rb)
}

func (c FfiConverterTypeTwoResourcePoolContributionEvent) Read(reader io.Reader) TwoResourcePoolContributionEvent {
	return TwoResourcePoolContributionEvent {
			FfiConverterMapStringDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTwoResourcePoolContributionEvent) Lower(value TwoResourcePoolContributionEvent) RustBuffer {
	return LowerIntoRustBuffer[TwoResourcePoolContributionEvent](c, value)
}

func (c FfiConverterTypeTwoResourcePoolContributionEvent) Write(writer io.Writer, value TwoResourcePoolContributionEvent) {
		FfiConverterMapStringDecimalINSTANCE.Write(writer, value.ContributedResources);
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitsMinted);
}

type FfiDestroyerTypeTwoResourcePoolContributionEvent struct {}

func (_ FfiDestroyerTypeTwoResourcePoolContributionEvent) Destroy(value TwoResourcePoolContributionEvent) {
	value.Destroy()
}


type TwoResourcePoolDepositEvent struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (r *TwoResourcePoolDepositEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeTwoResourcePoolDepositEvent struct {}

var FfiConverterTypeTwoResourcePoolDepositEventINSTANCE = FfiConverterTypeTwoResourcePoolDepositEvent{}

func (c FfiConverterTypeTwoResourcePoolDepositEvent) Lift(rb RustBufferI) TwoResourcePoolDepositEvent {
	return LiftFromRustBuffer[TwoResourcePoolDepositEvent](c, rb)
}

func (c FfiConverterTypeTwoResourcePoolDepositEvent) Read(reader io.Reader) TwoResourcePoolDepositEvent {
	return TwoResourcePoolDepositEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTwoResourcePoolDepositEvent) Lower(value TwoResourcePoolDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[TwoResourcePoolDepositEvent](c, value)
}

func (c FfiConverterTypeTwoResourcePoolDepositEvent) Write(writer io.Writer, value TwoResourcePoolDepositEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeTwoResourcePoolDepositEvent struct {}

func (_ FfiDestroyerTypeTwoResourcePoolDepositEvent) Destroy(value TwoResourcePoolDepositEvent) {
	value.Destroy()
}


type TwoResourcePoolRedemptionEvent struct {
	PoolUnitTokensRedeemed *Decimal
	RedeemedResources map[string]*Decimal
}

func (r *TwoResourcePoolRedemptionEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.PoolUnitTokensRedeemed);
		FfiDestroyerMapStringDecimal{}.Destroy(r.RedeemedResources);
}

type FfiConverterTypeTwoResourcePoolRedemptionEvent struct {}

var FfiConverterTypeTwoResourcePoolRedemptionEventINSTANCE = FfiConverterTypeTwoResourcePoolRedemptionEvent{}

func (c FfiConverterTypeTwoResourcePoolRedemptionEvent) Lift(rb RustBufferI) TwoResourcePoolRedemptionEvent {
	return LiftFromRustBuffer[TwoResourcePoolRedemptionEvent](c, rb)
}

func (c FfiConverterTypeTwoResourcePoolRedemptionEvent) Read(reader io.Reader) TwoResourcePoolRedemptionEvent {
	return TwoResourcePoolRedemptionEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterMapStringDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTwoResourcePoolRedemptionEvent) Lower(value TwoResourcePoolRedemptionEvent) RustBuffer {
	return LowerIntoRustBuffer[TwoResourcePoolRedemptionEvent](c, value)
}

func (c FfiConverterTypeTwoResourcePoolRedemptionEvent) Write(writer io.Writer, value TwoResourcePoolRedemptionEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.PoolUnitTokensRedeemed);
		FfiConverterMapStringDecimalINSTANCE.Write(writer, value.RedeemedResources);
}

type FfiDestroyerTypeTwoResourcePoolRedemptionEvent struct {}

func (_ FfiDestroyerTypeTwoResourcePoolRedemptionEvent) Destroy(value TwoResourcePoolRedemptionEvent) {
	value.Destroy()
}


type TwoResourcePoolWithdrawEvent struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (r *TwoResourcePoolWithdrawEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeTwoResourcePoolWithdrawEvent struct {}

var FfiConverterTypeTwoResourcePoolWithdrawEventINSTANCE = FfiConverterTypeTwoResourcePoolWithdrawEvent{}

func (c FfiConverterTypeTwoResourcePoolWithdrawEvent) Lift(rb RustBufferI) TwoResourcePoolWithdrawEvent {
	return LiftFromRustBuffer[TwoResourcePoolWithdrawEvent](c, rb)
}

func (c FfiConverterTypeTwoResourcePoolWithdrawEvent) Read(reader io.Reader) TwoResourcePoolWithdrawEvent {
	return TwoResourcePoolWithdrawEvent {
			FfiConverterAddressINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTwoResourcePoolWithdrawEvent) Lower(value TwoResourcePoolWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[TwoResourcePoolWithdrawEvent](c, value)
}

func (c FfiConverterTypeTwoResourcePoolWithdrawEvent) Write(writer io.Writer, value TwoResourcePoolWithdrawEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.ResourceAddress);
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeTwoResourcePoolWithdrawEvent struct {}

func (_ FfiDestroyerTypeTwoResourcePoolWithdrawEvent) Destroy(value TwoResourcePoolWithdrawEvent) {
	value.Destroy()
}


type UnlockPrimaryRoleEvent struct {
	PlaceholderField bool
}

func (r *UnlockPrimaryRoleEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.PlaceholderField);
}

type FfiConverterTypeUnlockPrimaryRoleEvent struct {}

var FfiConverterTypeUnlockPrimaryRoleEventINSTANCE = FfiConverterTypeUnlockPrimaryRoleEvent{}

func (c FfiConverterTypeUnlockPrimaryRoleEvent) Lift(rb RustBufferI) UnlockPrimaryRoleEvent {
	return LiftFromRustBuffer[UnlockPrimaryRoleEvent](c, rb)
}

func (c FfiConverterTypeUnlockPrimaryRoleEvent) Read(reader io.Reader) UnlockPrimaryRoleEvent {
	return UnlockPrimaryRoleEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnlockPrimaryRoleEvent) Lower(value UnlockPrimaryRoleEvent) RustBuffer {
	return LowerIntoRustBuffer[UnlockPrimaryRoleEvent](c, value)
}

func (c FfiConverterTypeUnlockPrimaryRoleEvent) Write(writer io.Writer, value UnlockPrimaryRoleEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.PlaceholderField);
}

type FfiDestroyerTypeUnlockPrimaryRoleEvent struct {}

func (_ FfiDestroyerTypeUnlockPrimaryRoleEvent) Destroy(value UnlockPrimaryRoleEvent) {
	value.Destroy()
}


type UnregisterValidatorEvent struct {
	PlaceholderField bool
}

func (r *UnregisterValidatorEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.PlaceholderField);
}

type FfiConverterTypeUnregisterValidatorEvent struct {}

var FfiConverterTypeUnregisterValidatorEventINSTANCE = FfiConverterTypeUnregisterValidatorEvent{}

func (c FfiConverterTypeUnregisterValidatorEvent) Lift(rb RustBufferI) UnregisterValidatorEvent {
	return LiftFromRustBuffer[UnregisterValidatorEvent](c, rb)
}

func (c FfiConverterTypeUnregisterValidatorEvent) Read(reader io.Reader) UnregisterValidatorEvent {
	return UnregisterValidatorEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnregisterValidatorEvent) Lower(value UnregisterValidatorEvent) RustBuffer {
	return LowerIntoRustBuffer[UnregisterValidatorEvent](c, value)
}

func (c FfiConverterTypeUnregisterValidatorEvent) Write(writer io.Writer, value UnregisterValidatorEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.PlaceholderField);
}

type FfiDestroyerTypeUnregisterValidatorEvent struct {}

func (_ FfiDestroyerTypeUnregisterValidatorEvent) Destroy(value UnregisterValidatorEvent) {
	value.Destroy()
}


type UnstakeData struct {
	Name string
	ClaimEpoch uint64
	ClaimAmount *Decimal
}

func (r *UnstakeData) Destroy() {
		FfiDestroyerString{}.Destroy(r.Name);
		FfiDestroyerUint64{}.Destroy(r.ClaimEpoch);
		FfiDestroyerDecimal{}.Destroy(r.ClaimAmount);
}

type FfiConverterTypeUnstakeData struct {}

var FfiConverterTypeUnstakeDataINSTANCE = FfiConverterTypeUnstakeData{}

func (c FfiConverterTypeUnstakeData) Lift(rb RustBufferI) UnstakeData {
	return LiftFromRustBuffer[UnstakeData](c, rb)
}

func (c FfiConverterTypeUnstakeData) Read(reader io.Reader) UnstakeData {
	return UnstakeData {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnstakeData) Lower(value UnstakeData) RustBuffer {
	return LowerIntoRustBuffer[UnstakeData](c, value)
}

func (c FfiConverterTypeUnstakeData) Write(writer io.Writer, value UnstakeData) {
		FfiConverterStringINSTANCE.Write(writer, value.Name);
		FfiConverterUint64INSTANCE.Write(writer, value.ClaimEpoch);
		FfiConverterDecimalINSTANCE.Write(writer, value.ClaimAmount);
}

type FfiDestroyerTypeUnstakeData struct {}

func (_ FfiDestroyerTypeUnstakeData) Destroy(value UnstakeData) {
	value.Destroy()
}


type UnstakeDataEntry struct {
	NonFungibleGlobalId *NonFungibleGlobalId
	Data UnstakeData
}

func (r *UnstakeDataEntry) Destroy() {
		FfiDestroyerNonFungibleGlobalId{}.Destroy(r.NonFungibleGlobalId);
		FfiDestroyerTypeUnstakeData{}.Destroy(r.Data);
}

type FfiConverterTypeUnstakeDataEntry struct {}

var FfiConverterTypeUnstakeDataEntryINSTANCE = FfiConverterTypeUnstakeDataEntry{}

func (c FfiConverterTypeUnstakeDataEntry) Lift(rb RustBufferI) UnstakeDataEntry {
	return LiftFromRustBuffer[UnstakeDataEntry](c, rb)
}

func (c FfiConverterTypeUnstakeDataEntry) Read(reader io.Reader) UnstakeDataEntry {
	return UnstakeDataEntry {
			FfiConverterNonFungibleGlobalIdINSTANCE.Read(reader),
			FfiConverterTypeUnstakeDataINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnstakeDataEntry) Lower(value UnstakeDataEntry) RustBuffer {
	return LowerIntoRustBuffer[UnstakeDataEntry](c, value)
}

func (c FfiConverterTypeUnstakeDataEntry) Write(writer io.Writer, value UnstakeDataEntry) {
		FfiConverterNonFungibleGlobalIdINSTANCE.Write(writer, value.NonFungibleGlobalId);
		FfiConverterTypeUnstakeDataINSTANCE.Write(writer, value.Data);
}

type FfiDestroyerTypeUnstakeDataEntry struct {}

func (_ FfiDestroyerTypeUnstakeDataEntry) Destroy(value UnstakeDataEntry) {
	value.Destroy()
}


type UnstakeEvent struct {
	StakeUnits *Decimal
}

func (r *UnstakeEvent) Destroy() {
		FfiDestroyerDecimal{}.Destroy(r.StakeUnits);
}

type FfiConverterTypeUnstakeEvent struct {}

var FfiConverterTypeUnstakeEventINSTANCE = FfiConverterTypeUnstakeEvent{}

func (c FfiConverterTypeUnstakeEvent) Lift(rb RustBufferI) UnstakeEvent {
	return LiftFromRustBuffer[UnstakeEvent](c, rb)
}

func (c FfiConverterTypeUnstakeEvent) Read(reader io.Reader) UnstakeEvent {
	return UnstakeEvent {
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnstakeEvent) Lower(value UnstakeEvent) RustBuffer {
	return LowerIntoRustBuffer[UnstakeEvent](c, value)
}

func (c FfiConverterTypeUnstakeEvent) Write(writer io.Writer, value UnstakeEvent) {
		FfiConverterDecimalINSTANCE.Write(writer, value.StakeUnits);
}

type FfiDestroyerTypeUnstakeEvent struct {}

func (_ FfiDestroyerTypeUnstakeEvent) Destroy(value UnstakeEvent) {
	value.Destroy()
}


type UpdateAcceptingStakeDelegationStateEvent struct {
	AcceptsDelegation bool
}

func (r *UpdateAcceptingStakeDelegationStateEvent) Destroy() {
		FfiDestroyerBool{}.Destroy(r.AcceptsDelegation);
}

type FfiConverterTypeUpdateAcceptingStakeDelegationStateEvent struct {}

var FfiConverterTypeUpdateAcceptingStakeDelegationStateEventINSTANCE = FfiConverterTypeUpdateAcceptingStakeDelegationStateEvent{}

func (c FfiConverterTypeUpdateAcceptingStakeDelegationStateEvent) Lift(rb RustBufferI) UpdateAcceptingStakeDelegationStateEvent {
	return LiftFromRustBuffer[UpdateAcceptingStakeDelegationStateEvent](c, rb)
}

func (c FfiConverterTypeUpdateAcceptingStakeDelegationStateEvent) Read(reader io.Reader) UpdateAcceptingStakeDelegationStateEvent {
	return UpdateAcceptingStakeDelegationStateEvent {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUpdateAcceptingStakeDelegationStateEvent) Lower(value UpdateAcceptingStakeDelegationStateEvent) RustBuffer {
	return LowerIntoRustBuffer[UpdateAcceptingStakeDelegationStateEvent](c, value)
}

func (c FfiConverterTypeUpdateAcceptingStakeDelegationStateEvent) Write(writer io.Writer, value UpdateAcceptingStakeDelegationStateEvent) {
		FfiConverterBoolINSTANCE.Write(writer, value.AcceptsDelegation);
}

type FfiDestroyerTypeUpdateAcceptingStakeDelegationStateEvent struct {}

func (_ FfiDestroyerTypeUpdateAcceptingStakeDelegationStateEvent) Destroy(value UpdateAcceptingStakeDelegationStateEvent) {
	value.Destroy()
}


type ValidatorEmissionAppliedEvent struct {
	Epoch uint64
	StartingStakePoolXrd *Decimal
	StakePoolAddedXrd *Decimal
	TotalStakeUnitSupply *Decimal
	ValidatorFeeXrd *Decimal
	ProposalsMade uint64
	ProposalsMissed uint64
}

func (r *ValidatorEmissionAppliedEvent) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.Epoch);
		FfiDestroyerDecimal{}.Destroy(r.StartingStakePoolXrd);
		FfiDestroyerDecimal{}.Destroy(r.StakePoolAddedXrd);
		FfiDestroyerDecimal{}.Destroy(r.TotalStakeUnitSupply);
		FfiDestroyerDecimal{}.Destroy(r.ValidatorFeeXrd);
		FfiDestroyerUint64{}.Destroy(r.ProposalsMade);
		FfiDestroyerUint64{}.Destroy(r.ProposalsMissed);
}

type FfiConverterTypeValidatorEmissionAppliedEvent struct {}

var FfiConverterTypeValidatorEmissionAppliedEventINSTANCE = FfiConverterTypeValidatorEmissionAppliedEvent{}

func (c FfiConverterTypeValidatorEmissionAppliedEvent) Lift(rb RustBufferI) ValidatorEmissionAppliedEvent {
	return LiftFromRustBuffer[ValidatorEmissionAppliedEvent](c, rb)
}

func (c FfiConverterTypeValidatorEmissionAppliedEvent) Read(reader io.Reader) ValidatorEmissionAppliedEvent {
	return ValidatorEmissionAppliedEvent {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeValidatorEmissionAppliedEvent) Lower(value ValidatorEmissionAppliedEvent) RustBuffer {
	return LowerIntoRustBuffer[ValidatorEmissionAppliedEvent](c, value)
}

func (c FfiConverterTypeValidatorEmissionAppliedEvent) Write(writer io.Writer, value ValidatorEmissionAppliedEvent) {
		FfiConverterUint64INSTANCE.Write(writer, value.Epoch);
		FfiConverterDecimalINSTANCE.Write(writer, value.StartingStakePoolXrd);
		FfiConverterDecimalINSTANCE.Write(writer, value.StakePoolAddedXrd);
		FfiConverterDecimalINSTANCE.Write(writer, value.TotalStakeUnitSupply);
		FfiConverterDecimalINSTANCE.Write(writer, value.ValidatorFeeXrd);
		FfiConverterUint64INSTANCE.Write(writer, value.ProposalsMade);
		FfiConverterUint64INSTANCE.Write(writer, value.ProposalsMissed);
}

type FfiDestroyerTypeValidatorEmissionAppliedEvent struct {}

func (_ FfiDestroyerTypeValidatorEmissionAppliedEvent) Destroy(value ValidatorEmissionAppliedEvent) {
	value.Destroy()
}


type ValidatorInfo struct {
	Key Secp256k1PublicKey
	Stake *Decimal
}

func (r *ValidatorInfo) Destroy() {
		FfiDestroyerTypeSecp256k1PublicKey{}.Destroy(r.Key);
		FfiDestroyerDecimal{}.Destroy(r.Stake);
}

type FfiConverterTypeValidatorInfo struct {}

var FfiConverterTypeValidatorInfoINSTANCE = FfiConverterTypeValidatorInfo{}

func (c FfiConverterTypeValidatorInfo) Lift(rb RustBufferI) ValidatorInfo {
	return LiftFromRustBuffer[ValidatorInfo](c, rb)
}

func (c FfiConverterTypeValidatorInfo) Read(reader io.Reader) ValidatorInfo {
	return ValidatorInfo {
			FfiConverterTypeSecp256k1PublicKeyINSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeValidatorInfo) Lower(value ValidatorInfo) RustBuffer {
	return LowerIntoRustBuffer[ValidatorInfo](c, value)
}

func (c FfiConverterTypeValidatorInfo) Write(writer io.Writer, value ValidatorInfo) {
		FfiConverterTypeSecp256k1PublicKeyINSTANCE.Write(writer, value.Key);
		FfiConverterDecimalINSTANCE.Write(writer, value.Stake);
}

type FfiDestroyerTypeValidatorInfo struct {}

func (_ FfiDestroyerTypeValidatorInfo) Destroy(value ValidatorInfo) {
	value.Destroy()
}


type ValidatorRewardAppliedEvent struct {
	Epoch uint64
	Amount *Decimal
}

func (r *ValidatorRewardAppliedEvent) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.Epoch);
		FfiDestroyerDecimal{}.Destroy(r.Amount);
}

type FfiConverterTypeValidatorRewardAppliedEvent struct {}

var FfiConverterTypeValidatorRewardAppliedEventINSTANCE = FfiConverterTypeValidatorRewardAppliedEvent{}

func (c FfiConverterTypeValidatorRewardAppliedEvent) Lift(rb RustBufferI) ValidatorRewardAppliedEvent {
	return LiftFromRustBuffer[ValidatorRewardAppliedEvent](c, rb)
}

func (c FfiConverterTypeValidatorRewardAppliedEvent) Read(reader io.Reader) ValidatorRewardAppliedEvent {
	return ValidatorRewardAppliedEvent {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterDecimalINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeValidatorRewardAppliedEvent) Lower(value ValidatorRewardAppliedEvent) RustBuffer {
	return LowerIntoRustBuffer[ValidatorRewardAppliedEvent](c, value)
}

func (c FfiConverterTypeValidatorRewardAppliedEvent) Write(writer io.Writer, value ValidatorRewardAppliedEvent) {
		FfiConverterUint64INSTANCE.Write(writer, value.Epoch);
		FfiConverterDecimalINSTANCE.Write(writer, value.Amount);
}

type FfiDestroyerTypeValidatorRewardAppliedEvent struct {}

func (_ FfiDestroyerTypeValidatorRewardAppliedEvent) Destroy(value ValidatorRewardAppliedEvent) {
	value.Destroy()
}


type VaultCreationEvent struct {
	VaultId *Address
}

func (r *VaultCreationEvent) Destroy() {
		FfiDestroyerAddress{}.Destroy(r.VaultId);
}

type FfiConverterTypeVaultCreationEvent struct {}

var FfiConverterTypeVaultCreationEventINSTANCE = FfiConverterTypeVaultCreationEvent{}

func (c FfiConverterTypeVaultCreationEvent) Lift(rb RustBufferI) VaultCreationEvent {
	return LiftFromRustBuffer[VaultCreationEvent](c, rb)
}

func (c FfiConverterTypeVaultCreationEvent) Read(reader io.Reader) VaultCreationEvent {
	return VaultCreationEvent {
			FfiConverterAddressINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeVaultCreationEvent) Lower(value VaultCreationEvent) RustBuffer {
	return LowerIntoRustBuffer[VaultCreationEvent](c, value)
}

func (c FfiConverterTypeVaultCreationEvent) Write(writer io.Writer, value VaultCreationEvent) {
		FfiConverterAddressINSTANCE.Write(writer, value.VaultId);
}

type FfiDestroyerTypeVaultCreationEvent struct {}

func (_ FfiDestroyerTypeVaultCreationEvent) Destroy(value VaultCreationEvent) {
	value.Destroy()
}



type AccountDefaultDepositRule uint

const (
	AccountDefaultDepositRuleAccept AccountDefaultDepositRule = 1
	AccountDefaultDepositRuleReject AccountDefaultDepositRule = 2
	AccountDefaultDepositRuleAllowExisting AccountDefaultDepositRule = 3
)

type FfiConverterTypeAccountDefaultDepositRule struct {}

var FfiConverterTypeAccountDefaultDepositRuleINSTANCE = FfiConverterTypeAccountDefaultDepositRule{}

func (c FfiConverterTypeAccountDefaultDepositRule) Lift(rb RustBufferI) AccountDefaultDepositRule {
	return LiftFromRustBuffer[AccountDefaultDepositRule](c, rb)
}

func (c FfiConverterTypeAccountDefaultDepositRule) Lower(value AccountDefaultDepositRule) RustBuffer {
	return LowerIntoRustBuffer[AccountDefaultDepositRule](c, value)
}
func (FfiConverterTypeAccountDefaultDepositRule) Read(reader io.Reader) AccountDefaultDepositRule {
	id := readInt32(reader)
	return AccountDefaultDepositRule(id)
}

func (FfiConverterTypeAccountDefaultDepositRule) Write(writer io.Writer, value AccountDefaultDepositRule) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeAccountDefaultDepositRule struct {}

func (_ FfiDestroyerTypeAccountDefaultDepositRule) Destroy(value AccountDefaultDepositRule) {
}




type AccountDepositEvent interface {
	Destroy()
}
type AccountDepositEventFungible struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e AccountDepositEventFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type AccountDepositEventNonFungible struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e AccountDepositEventNonFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}

type FfiConverterTypeAccountDepositEvent struct {}

var FfiConverterTypeAccountDepositEventINSTANCE = FfiConverterTypeAccountDepositEvent{}

func (c FfiConverterTypeAccountDepositEvent) Lift(rb RustBufferI) AccountDepositEvent {
	return LiftFromRustBuffer[AccountDepositEvent](c, rb)
}

func (c FfiConverterTypeAccountDepositEvent) Lower(value AccountDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountDepositEvent](c, value)
}
func (FfiConverterTypeAccountDepositEvent) Read(reader io.Reader) AccountDepositEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return AccountDepositEventFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return AccountDepositEventNonFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeAccountDepositEvent.Read()", id));
	}
}

func (FfiConverterTypeAccountDepositEvent) Write(writer io.Writer, value AccountDepositEvent) {
	switch variant_value := value.(type) {
		case AccountDepositEventFungible:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case AccountDepositEventNonFungible:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeAccountDepositEvent.Write", value))
	}
}

type FfiDestroyerTypeAccountDepositEvent struct {}

func (_ FfiDestroyerTypeAccountDepositEvent) Destroy(value AccountDepositEvent) {
	value.Destroy()
}




type AccountRejectedDepositEvent interface {
	Destroy()
}
type AccountRejectedDepositEventFungible struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e AccountRejectedDepositEventFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type AccountRejectedDepositEventNonFungible struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e AccountRejectedDepositEventNonFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}

type FfiConverterTypeAccountRejectedDepositEvent struct {}

var FfiConverterTypeAccountRejectedDepositEventINSTANCE = FfiConverterTypeAccountRejectedDepositEvent{}

func (c FfiConverterTypeAccountRejectedDepositEvent) Lift(rb RustBufferI) AccountRejectedDepositEvent {
	return LiftFromRustBuffer[AccountRejectedDepositEvent](c, rb)
}

func (c FfiConverterTypeAccountRejectedDepositEvent) Lower(value AccountRejectedDepositEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountRejectedDepositEvent](c, value)
}
func (FfiConverterTypeAccountRejectedDepositEvent) Read(reader io.Reader) AccountRejectedDepositEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return AccountRejectedDepositEventFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return AccountRejectedDepositEventNonFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeAccountRejectedDepositEvent.Read()", id));
	}
}

func (FfiConverterTypeAccountRejectedDepositEvent) Write(writer io.Writer, value AccountRejectedDepositEvent) {
	switch variant_value := value.(type) {
		case AccountRejectedDepositEventFungible:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case AccountRejectedDepositEventNonFungible:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeAccountRejectedDepositEvent.Write", value))
	}
}

type FfiDestroyerTypeAccountRejectedDepositEvent struct {}

func (_ FfiDestroyerTypeAccountRejectedDepositEvent) Destroy(value AccountRejectedDepositEvent) {
	value.Destroy()
}




type AccountWithdrawEvent interface {
	Destroy()
}
type AccountWithdrawEventFungible struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e AccountWithdrawEventFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type AccountWithdrawEventNonFungible struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e AccountWithdrawEventNonFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}

type FfiConverterTypeAccountWithdrawEvent struct {}

var FfiConverterTypeAccountWithdrawEventINSTANCE = FfiConverterTypeAccountWithdrawEvent{}

func (c FfiConverterTypeAccountWithdrawEvent) Lift(rb RustBufferI) AccountWithdrawEvent {
	return LiftFromRustBuffer[AccountWithdrawEvent](c, rb)
}

func (c FfiConverterTypeAccountWithdrawEvent) Lower(value AccountWithdrawEvent) RustBuffer {
	return LowerIntoRustBuffer[AccountWithdrawEvent](c, value)
}
func (FfiConverterTypeAccountWithdrawEvent) Read(reader io.Reader) AccountWithdrawEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return AccountWithdrawEventFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return AccountWithdrawEventNonFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeAccountWithdrawEvent.Read()", id));
	}
}

func (FfiConverterTypeAccountWithdrawEvent) Write(writer io.Writer, value AccountWithdrawEvent) {
	switch variant_value := value.(type) {
		case AccountWithdrawEventFungible:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case AccountWithdrawEventNonFungible:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeAccountWithdrawEvent.Write", value))
	}
}

type FfiDestroyerTypeAccountWithdrawEvent struct {}

func (_ FfiDestroyerTypeAccountWithdrawEvent) Destroy(value AccountWithdrawEvent) {
	value.Destroy()
}




type Assertion interface {
	Destroy()
}
type AssertionAmount struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e AssertionAmount) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type AssertionIds struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e AssertionIds) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}

type FfiConverterTypeAssertion struct {}

var FfiConverterTypeAssertionINSTANCE = FfiConverterTypeAssertion{}

func (c FfiConverterTypeAssertion) Lift(rb RustBufferI) Assertion {
	return LiftFromRustBuffer[Assertion](c, rb)
}

func (c FfiConverterTypeAssertion) Lower(value Assertion) RustBuffer {
	return LowerIntoRustBuffer[Assertion](c, value)
}
func (FfiConverterTypeAssertion) Read(reader io.Reader) Assertion {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return AssertionAmount{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return AssertionIds{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeAssertion.Read()", id));
	}
}

func (FfiConverterTypeAssertion) Write(writer io.Writer, value Assertion) {
	switch variant_value := value.(type) {
		case AssertionAmount:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case AssertionIds:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeAssertion.Write", value))
	}
}

type FfiDestroyerTypeAssertion struct {}

func (_ FfiDestroyerTypeAssertion) Destroy(value Assertion) {
	value.Destroy()
}




type Curve uint

const (
	CurveSecp256k1 Curve = 1
	CurveEd25519 Curve = 2
)

type FfiConverterTypeCurve struct {}

var FfiConverterTypeCurveINSTANCE = FfiConverterTypeCurve{}

func (c FfiConverterTypeCurve) Lift(rb RustBufferI) Curve {
	return LiftFromRustBuffer[Curve](c, rb)
}

func (c FfiConverterTypeCurve) Lower(value Curve) RustBuffer {
	return LowerIntoRustBuffer[Curve](c, value)
}
func (FfiConverterTypeCurve) Read(reader io.Reader) Curve {
	id := readInt32(reader)
	return Curve(id)
}

func (FfiConverterTypeCurve) Write(writer io.Writer, value Curve) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeCurve struct {}

func (_ FfiDestroyerTypeCurve) Destroy(value Curve) {
}




type CurveType uint

const (
	CurveTypeEd25519 CurveType = 1
	CurveTypeSecp256k1 CurveType = 2
)

type FfiConverterTypeCurveType struct {}

var FfiConverterTypeCurveTypeINSTANCE = FfiConverterTypeCurveType{}

func (c FfiConverterTypeCurveType) Lift(rb RustBufferI) CurveType {
	return LiftFromRustBuffer[CurveType](c, rb)
}

func (c FfiConverterTypeCurveType) Lower(value CurveType) RustBuffer {
	return LowerIntoRustBuffer[CurveType](c, value)
}
func (FfiConverterTypeCurveType) Read(reader io.Reader) CurveType {
	id := readInt32(reader)
	return CurveType(id)
}

func (FfiConverterTypeCurveType) Write(writer io.Writer, value CurveType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeCurveType struct {}

func (_ FfiDestroyerTypeCurveType) Destroy(value CurveType) {
}




type DecryptorsByCurve interface {
	Destroy()
}
type DecryptorsByCurveEd25519 struct {
	DhEphemeralPublicKey Ed25519PublicKey
	Decryptors map[PublicKeyFingerprint][]byte
}

func (e DecryptorsByCurveEd25519) Destroy() {
		FfiDestroyerTypeEd25519PublicKey{}.Destroy(e.DhEphemeralPublicKey);
		FfiDestroyerMapTypePublicKeyFingerprintBytes{}.Destroy(e.Decryptors);
}
type DecryptorsByCurveSecp256k1 struct {
	DhEphemeralPublicKey Secp256k1PublicKey
	Decryptors map[PublicKeyFingerprint][]byte
}

func (e DecryptorsByCurveSecp256k1) Destroy() {
		FfiDestroyerTypeSecp256k1PublicKey{}.Destroy(e.DhEphemeralPublicKey);
		FfiDestroyerMapTypePublicKeyFingerprintBytes{}.Destroy(e.Decryptors);
}

type FfiConverterTypeDecryptorsByCurve struct {}

var FfiConverterTypeDecryptorsByCurveINSTANCE = FfiConverterTypeDecryptorsByCurve{}

func (c FfiConverterTypeDecryptorsByCurve) Lift(rb RustBufferI) DecryptorsByCurve {
	return LiftFromRustBuffer[DecryptorsByCurve](c, rb)
}

func (c FfiConverterTypeDecryptorsByCurve) Lower(value DecryptorsByCurve) RustBuffer {
	return LowerIntoRustBuffer[DecryptorsByCurve](c, value)
}
func (FfiConverterTypeDecryptorsByCurve) Read(reader io.Reader) DecryptorsByCurve {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return DecryptorsByCurveEd25519{
				FfiConverterTypeEd25519PublicKeyINSTANCE.Read(reader),
				FfiConverterMapTypePublicKeyFingerprintBytesINSTANCE.Read(reader),
			};
		case 2:
			return DecryptorsByCurveSecp256k1{
				FfiConverterTypeSecp256k1PublicKeyINSTANCE.Read(reader),
				FfiConverterMapTypePublicKeyFingerprintBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeDecryptorsByCurve.Read()", id));
	}
}

func (FfiConverterTypeDecryptorsByCurve) Write(writer io.Writer, value DecryptorsByCurve) {
	switch variant_value := value.(type) {
		case DecryptorsByCurveEd25519:
			writeInt32(writer, 1)
			FfiConverterTypeEd25519PublicKeyINSTANCE.Write(writer, variant_value.DhEphemeralPublicKey)
			FfiConverterMapTypePublicKeyFingerprintBytesINSTANCE.Write(writer, variant_value.Decryptors)
		case DecryptorsByCurveSecp256k1:
			writeInt32(writer, 2)
			FfiConverterTypeSecp256k1PublicKeyINSTANCE.Write(writer, variant_value.DhEphemeralPublicKey)
			FfiConverterMapTypePublicKeyFingerprintBytesINSTANCE.Write(writer, variant_value.Decryptors)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeDecryptorsByCurve.Write", value))
	}
}

type FfiDestroyerTypeDecryptorsByCurve struct {}

func (_ FfiDestroyerTypeDecryptorsByCurve) Destroy(value DecryptorsByCurve) {
	value.Destroy()
}




type DependencyInformation interface {
	Destroy()
}
type DependencyInformationVersion struct {
	Value string
}

func (e DependencyInformationVersion) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type DependencyInformationTag struct {
	Value string
}

func (e DependencyInformationTag) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type DependencyInformationBranch struct {
	Value string
}

func (e DependencyInformationBranch) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type DependencyInformationRev struct {
	Value string
}

func (e DependencyInformationRev) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}

type FfiConverterTypeDependencyInformation struct {}

var FfiConverterTypeDependencyInformationINSTANCE = FfiConverterTypeDependencyInformation{}

func (c FfiConverterTypeDependencyInformation) Lift(rb RustBufferI) DependencyInformation {
	return LiftFromRustBuffer[DependencyInformation](c, rb)
}

func (c FfiConverterTypeDependencyInformation) Lower(value DependencyInformation) RustBuffer {
	return LowerIntoRustBuffer[DependencyInformation](c, value)
}
func (FfiConverterTypeDependencyInformation) Read(reader io.Reader) DependencyInformation {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return DependencyInformationVersion{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return DependencyInformationTag{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 3:
			return DependencyInformationBranch{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 4:
			return DependencyInformationRev{
				FfiConverterStringINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeDependencyInformation.Read()", id));
	}
}

func (FfiConverterTypeDependencyInformation) Write(writer io.Writer, value DependencyInformation) {
	switch variant_value := value.(type) {
		case DependencyInformationVersion:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case DependencyInformationTag:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case DependencyInformationBranch:
			writeInt32(writer, 3)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case DependencyInformationRev:
			writeInt32(writer, 4)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeDependencyInformation.Write", value))
	}
}

type FfiDestroyerTypeDependencyInformation struct {}

func (_ FfiDestroyerTypeDependencyInformation) Destroy(value DependencyInformation) {
	value.Destroy()
}




type DepositResourceEvent interface {
	Destroy()
}
type DepositResourceEventAmount struct {
	Value *Decimal
}

func (e DepositResourceEventAmount) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type DepositResourceEventIds struct {
	Value []NonFungibleLocalId
}

func (e DepositResourceEventIds) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Value);
}

type FfiConverterTypeDepositResourceEvent struct {}

var FfiConverterTypeDepositResourceEventINSTANCE = FfiConverterTypeDepositResourceEvent{}

func (c FfiConverterTypeDepositResourceEvent) Lift(rb RustBufferI) DepositResourceEvent {
	return LiftFromRustBuffer[DepositResourceEvent](c, rb)
}

func (c FfiConverterTypeDepositResourceEvent) Lower(value DepositResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[DepositResourceEvent](c, value)
}
func (FfiConverterTypeDepositResourceEvent) Read(reader io.Reader) DepositResourceEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return DepositResourceEventAmount{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return DepositResourceEventIds{
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeDepositResourceEvent.Read()", id));
	}
}

func (FfiConverterTypeDepositResourceEvent) Write(writer io.Writer, value DepositResourceEvent) {
	switch variant_value := value.(type) {
		case DepositResourceEventAmount:
			writeInt32(writer, 1)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case DepositResourceEventIds:
			writeInt32(writer, 2)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeDepositResourceEvent.Write", value))
	}
}

type FfiDestroyerTypeDepositResourceEvent struct {}

func (_ FfiDestroyerTypeDepositResourceEvent) Destroy(value DepositResourceEvent) {
	value.Destroy()
}




type DetailedManifestClass interface {
	Destroy()
}
type DetailedManifestClassGeneral struct {
}

func (e DetailedManifestClassGeneral) Destroy() {
}
type DetailedManifestClassTransfer struct {
	IsOneToOne bool
}

func (e DetailedManifestClassTransfer) Destroy() {
		FfiDestroyerBool{}.Destroy(e.IsOneToOne);
}
type DetailedManifestClassPoolContribution struct {
	PoolAddresses []*Address
	PoolContributions []TrackedPoolContribution
}

func (e DetailedManifestClassPoolContribution) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(e.PoolAddresses);
		FfiDestroyerSequenceTypeTrackedPoolContribution{}.Destroy(e.PoolContributions);
}
type DetailedManifestClassPoolRedemption struct {
	PoolAddresses []*Address
	PoolRedemptions []TrackedPoolRedemption
}

func (e DetailedManifestClassPoolRedemption) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(e.PoolAddresses);
		FfiDestroyerSequenceTypeTrackedPoolRedemption{}.Destroy(e.PoolRedemptions);
}
type DetailedManifestClassValidatorStake struct {
	ValidatorAddresses []*Address
	ValidatorStakes []TrackedValidatorStake
}

func (e DetailedManifestClassValidatorStake) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(e.ValidatorAddresses);
		FfiDestroyerSequenceTypeTrackedValidatorStake{}.Destroy(e.ValidatorStakes);
}
type DetailedManifestClassValidatorUnstake struct {
	ValidatorAddresses []*Address
	ValidatorUnstakes []TrackedValidatorUnstake
	ClaimsNonFungibleData []UnstakeDataEntry
}

func (e DetailedManifestClassValidatorUnstake) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(e.ValidatorAddresses);
		FfiDestroyerSequenceTypeTrackedValidatorUnstake{}.Destroy(e.ValidatorUnstakes);
		FfiDestroyerSequenceTypeUnstakeDataEntry{}.Destroy(e.ClaimsNonFungibleData);
}
type DetailedManifestClassValidatorClaim struct {
	ValidatorAddresses []*Address
	ValidatorClaims []TrackedValidatorClaim
}

func (e DetailedManifestClassValidatorClaim) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(e.ValidatorAddresses);
		FfiDestroyerSequenceTypeTrackedValidatorClaim{}.Destroy(e.ValidatorClaims);
}
type DetailedManifestClassAccountDepositSettingsUpdate struct {
	ResourcePreferencesUpdates map[string]map[string]ResourcePreferenceUpdate
	DepositModeUpdates map[string]AccountDefaultDepositRule
	AuthorizedDepositorsAdded map[string][]ResourceOrNonFungible
	AuthorizedDepositorsRemoved map[string][]ResourceOrNonFungible
}

func (e DetailedManifestClassAccountDepositSettingsUpdate) Destroy() {
		FfiDestroyerMapStringMapStringTypeResourcePreferenceUpdate{}.Destroy(e.ResourcePreferencesUpdates);
		FfiDestroyerMapStringTypeAccountDefaultDepositRule{}.Destroy(e.DepositModeUpdates);
		FfiDestroyerMapStringSequenceTypeResourceOrNonFungible{}.Destroy(e.AuthorizedDepositorsAdded);
		FfiDestroyerMapStringSequenceTypeResourceOrNonFungible{}.Destroy(e.AuthorizedDepositorsRemoved);
}

type FfiConverterTypeDetailedManifestClass struct {}

var FfiConverterTypeDetailedManifestClassINSTANCE = FfiConverterTypeDetailedManifestClass{}

func (c FfiConverterTypeDetailedManifestClass) Lift(rb RustBufferI) DetailedManifestClass {
	return LiftFromRustBuffer[DetailedManifestClass](c, rb)
}

func (c FfiConverterTypeDetailedManifestClass) Lower(value DetailedManifestClass) RustBuffer {
	return LowerIntoRustBuffer[DetailedManifestClass](c, value)
}
func (FfiConverterTypeDetailedManifestClass) Read(reader io.Reader) DetailedManifestClass {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return DetailedManifestClassGeneral{
			};
		case 2:
			return DetailedManifestClassTransfer{
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 3:
			return DetailedManifestClassPoolContribution{
				FfiConverterSequenceAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeTrackedPoolContributionINSTANCE.Read(reader),
			};
		case 4:
			return DetailedManifestClassPoolRedemption{
				FfiConverterSequenceAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeTrackedPoolRedemptionINSTANCE.Read(reader),
			};
		case 5:
			return DetailedManifestClassValidatorStake{
				FfiConverterSequenceAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeTrackedValidatorStakeINSTANCE.Read(reader),
			};
		case 6:
			return DetailedManifestClassValidatorUnstake{
				FfiConverterSequenceAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeTrackedValidatorUnstakeINSTANCE.Read(reader),
				FfiConverterSequenceTypeUnstakeDataEntryINSTANCE.Read(reader),
			};
		case 7:
			return DetailedManifestClassValidatorClaim{
				FfiConverterSequenceAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeTrackedValidatorClaimINSTANCE.Read(reader),
			};
		case 8:
			return DetailedManifestClassAccountDepositSettingsUpdate{
				FfiConverterMapStringMapStringTypeResourcePreferenceUpdateINSTANCE.Read(reader),
				FfiConverterMapStringTypeAccountDefaultDepositRuleINSTANCE.Read(reader),
				FfiConverterMapStringSequenceTypeResourceOrNonFungibleINSTANCE.Read(reader),
				FfiConverterMapStringSequenceTypeResourceOrNonFungibleINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeDetailedManifestClass.Read()", id));
	}
}

func (FfiConverterTypeDetailedManifestClass) Write(writer io.Writer, value DetailedManifestClass) {
	switch variant_value := value.(type) {
		case DetailedManifestClassGeneral:
			writeInt32(writer, 1)
		case DetailedManifestClassTransfer:
			writeInt32(writer, 2)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.IsOneToOne)
		case DetailedManifestClassPoolContribution:
			writeInt32(writer, 3)
			FfiConverterSequenceAddressINSTANCE.Write(writer, variant_value.PoolAddresses)
			FfiConverterSequenceTypeTrackedPoolContributionINSTANCE.Write(writer, variant_value.PoolContributions)
		case DetailedManifestClassPoolRedemption:
			writeInt32(writer, 4)
			FfiConverterSequenceAddressINSTANCE.Write(writer, variant_value.PoolAddresses)
			FfiConverterSequenceTypeTrackedPoolRedemptionINSTANCE.Write(writer, variant_value.PoolRedemptions)
		case DetailedManifestClassValidatorStake:
			writeInt32(writer, 5)
			FfiConverterSequenceAddressINSTANCE.Write(writer, variant_value.ValidatorAddresses)
			FfiConverterSequenceTypeTrackedValidatorStakeINSTANCE.Write(writer, variant_value.ValidatorStakes)
		case DetailedManifestClassValidatorUnstake:
			writeInt32(writer, 6)
			FfiConverterSequenceAddressINSTANCE.Write(writer, variant_value.ValidatorAddresses)
			FfiConverterSequenceTypeTrackedValidatorUnstakeINSTANCE.Write(writer, variant_value.ValidatorUnstakes)
			FfiConverterSequenceTypeUnstakeDataEntryINSTANCE.Write(writer, variant_value.ClaimsNonFungibleData)
		case DetailedManifestClassValidatorClaim:
			writeInt32(writer, 7)
			FfiConverterSequenceAddressINSTANCE.Write(writer, variant_value.ValidatorAddresses)
			FfiConverterSequenceTypeTrackedValidatorClaimINSTANCE.Write(writer, variant_value.ValidatorClaims)
		case DetailedManifestClassAccountDepositSettingsUpdate:
			writeInt32(writer, 8)
			FfiConverterMapStringMapStringTypeResourcePreferenceUpdateINSTANCE.Write(writer, variant_value.ResourcePreferencesUpdates)
			FfiConverterMapStringTypeAccountDefaultDepositRuleINSTANCE.Write(writer, variant_value.DepositModeUpdates)
			FfiConverterMapStringSequenceTypeResourceOrNonFungibleINSTANCE.Write(writer, variant_value.AuthorizedDepositorsAdded)
			FfiConverterMapStringSequenceTypeResourceOrNonFungibleINSTANCE.Write(writer, variant_value.AuthorizedDepositorsRemoved)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeDetailedManifestClass.Write", value))
	}
}

type FfiDestroyerTypeDetailedManifestClass struct {}

func (_ FfiDestroyerTypeDetailedManifestClass) Destroy(value DetailedManifestClass) {
	value.Destroy()
}




type Emitter interface {
	Destroy()
}
type EmitterFunction struct {
	Address *Address
	BlueprintName string
}

func (e EmitterFunction) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Address);
		FfiDestroyerString{}.Destroy(e.BlueprintName);
}
type EmitterMethod struct {
	Address *Address
	ObjectModuleId ModuleId
}

func (e EmitterMethod) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Address);
		FfiDestroyerTypeModuleId{}.Destroy(e.ObjectModuleId);
}

type FfiConverterTypeEmitter struct {}

var FfiConverterTypeEmitterINSTANCE = FfiConverterTypeEmitter{}

func (c FfiConverterTypeEmitter) Lift(rb RustBufferI) Emitter {
	return LiftFromRustBuffer[Emitter](c, rb)
}

func (c FfiConverterTypeEmitter) Lower(value Emitter) RustBuffer {
	return LowerIntoRustBuffer[Emitter](c, value)
}
func (FfiConverterTypeEmitter) Read(reader io.Reader) Emitter {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return EmitterFunction{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return EmitterMethod{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterTypeModuleIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeEmitter.Read()", id));
	}
}

func (FfiConverterTypeEmitter) Write(writer io.Writer, value Emitter) {
	switch variant_value := value.(type) {
		case EmitterFunction:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterStringINSTANCE.Write(writer, variant_value.BlueprintName)
		case EmitterMethod:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterTypeModuleIdINSTANCE.Write(writer, variant_value.ObjectModuleId)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeEmitter.Write", value))
	}
}

type FfiDestroyerTypeEmitter struct {}

func (_ FfiDestroyerTypeEmitter) Destroy(value Emitter) {
	value.Destroy()
}




type EntityType uint

const (
	EntityTypeGlobalPackage EntityType = 1
	EntityTypeGlobalFungibleResourceManager EntityType = 2
	EntityTypeGlobalNonFungibleResourceManager EntityType = 3
	EntityTypeGlobalConsensusManager EntityType = 4
	EntityTypeGlobalValidator EntityType = 5
	EntityTypeGlobalAccessController EntityType = 6
	EntityTypeGlobalAccount EntityType = 7
	EntityTypeGlobalIdentity EntityType = 8
	EntityTypeGlobalGenericComponent EntityType = 9
	EntityTypeGlobalVirtualSecp256k1Account EntityType = 10
	EntityTypeGlobalVirtualEd25519Account EntityType = 11
	EntityTypeGlobalVirtualSecp256k1Identity EntityType = 12
	EntityTypeGlobalVirtualEd25519Identity EntityType = 13
	EntityTypeGlobalOneResourcePool EntityType = 14
	EntityTypeGlobalTwoResourcePool EntityType = 15
	EntityTypeGlobalMultiResourcePool EntityType = 16
	EntityTypeGlobalAccountLocker EntityType = 17
	EntityTypeGlobalTransactionTracker EntityType = 18
	EntityTypeInternalFungibleVault EntityType = 19
	EntityTypeInternalNonFungibleVault EntityType = 20
	EntityTypeInternalGenericComponent EntityType = 21
	EntityTypeInternalKeyValueStore EntityType = 22
)

type FfiConverterTypeEntityType struct {}

var FfiConverterTypeEntityTypeINSTANCE = FfiConverterTypeEntityType{}

func (c FfiConverterTypeEntityType) Lift(rb RustBufferI) EntityType {
	return LiftFromRustBuffer[EntityType](c, rb)
}

func (c FfiConverterTypeEntityType) Lower(value EntityType) RustBuffer {
	return LowerIntoRustBuffer[EntityType](c, value)
}
func (FfiConverterTypeEntityType) Read(reader io.Reader) EntityType {
	id := readInt32(reader)
	return EntityType(id)
}

func (FfiConverterTypeEntityType) Write(writer io.Writer, value EntityType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeEntityType struct {}

func (_ FfiDestroyerTypeEntityType) Destroy(value EntityType) {
}




type FungibleResourceIndicator interface {
	Destroy()
}
type FungibleResourceIndicatorGuaranteed struct {
	Amount *Decimal
}

func (e FungibleResourceIndicatorGuaranteed) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type FungibleResourceIndicatorPredicted struct {
	PredictedAmount PredictedDecimal
}

func (e FungibleResourceIndicatorPredicted) Destroy() {
		FfiDestroyerTypePredictedDecimal{}.Destroy(e.PredictedAmount);
}

type FfiConverterTypeFungibleResourceIndicator struct {}

var FfiConverterTypeFungibleResourceIndicatorINSTANCE = FfiConverterTypeFungibleResourceIndicator{}

func (c FfiConverterTypeFungibleResourceIndicator) Lift(rb RustBufferI) FungibleResourceIndicator {
	return LiftFromRustBuffer[FungibleResourceIndicator](c, rb)
}

func (c FfiConverterTypeFungibleResourceIndicator) Lower(value FungibleResourceIndicator) RustBuffer {
	return LowerIntoRustBuffer[FungibleResourceIndicator](c, value)
}
func (FfiConverterTypeFungibleResourceIndicator) Read(reader io.Reader) FungibleResourceIndicator {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return FungibleResourceIndicatorGuaranteed{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return FungibleResourceIndicatorPredicted{
				FfiConverterTypePredictedDecimalINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeFungibleResourceIndicator.Read()", id));
	}
}

func (FfiConverterTypeFungibleResourceIndicator) Write(writer io.Writer, value FungibleResourceIndicator) {
	switch variant_value := value.(type) {
		case FungibleResourceIndicatorGuaranteed:
			writeInt32(writer, 1)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case FungibleResourceIndicatorPredicted:
			writeInt32(writer, 2)
			FfiConverterTypePredictedDecimalINSTANCE.Write(writer, variant_value.PredictedAmount)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeFungibleResourceIndicator.Write", value))
	}
}

type FfiDestroyerTypeFungibleResourceIndicator struct {}

func (_ FfiDestroyerTypeFungibleResourceIndicator) Destroy(value FungibleResourceIndicator) {
	value.Destroy()
}




type Instruction interface {
	Destroy()
}
type InstructionTakeAllFromWorktop struct {
	ResourceAddress *Address
}

func (e InstructionTakeAllFromWorktop) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
}
type InstructionTakeFromWorktop struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e InstructionTakeFromWorktop) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type InstructionTakeNonFungiblesFromWorktop struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e InstructionTakeNonFungiblesFromWorktop) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}
type InstructionReturnToWorktop struct {
	BucketId ManifestBucket
}

func (e InstructionReturnToWorktop) Destroy() {
		FfiDestroyerTypeManifestBucket{}.Destroy(e.BucketId);
}
type InstructionAssertWorktopContains struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e InstructionAssertWorktopContains) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type InstructionAssertWorktopContainsAny struct {
	ResourceAddress *Address
}

func (e InstructionAssertWorktopContainsAny) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
}
type InstructionAssertWorktopContainsNonFungibles struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e InstructionAssertWorktopContainsNonFungibles) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}
type InstructionPopFromAuthZone struct {
}

func (e InstructionPopFromAuthZone) Destroy() {
}
type InstructionPushToAuthZone struct {
	ProofId ManifestProof
}

func (e InstructionPushToAuthZone) Destroy() {
		FfiDestroyerTypeManifestProof{}.Destroy(e.ProofId);
}
type InstructionCreateProofFromAuthZoneOfAmount struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e InstructionCreateProofFromAuthZoneOfAmount) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type InstructionCreateProofFromAuthZoneOfNonFungibles struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e InstructionCreateProofFromAuthZoneOfNonFungibles) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}
type InstructionCreateProofFromAuthZoneOfAll struct {
	ResourceAddress *Address
}

func (e InstructionCreateProofFromAuthZoneOfAll) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
}
type InstructionDropAllProofs struct {
}

func (e InstructionDropAllProofs) Destroy() {
}
type InstructionDropNamedProofs struct {
}

func (e InstructionDropNamedProofs) Destroy() {
}
type InstructionDropAuthZoneProofs struct {
}

func (e InstructionDropAuthZoneProofs) Destroy() {
}
type InstructionDropAuthZoneRegularProofs struct {
}

func (e InstructionDropAuthZoneRegularProofs) Destroy() {
}
type InstructionDropAuthZoneSignatureProofs struct {
}

func (e InstructionDropAuthZoneSignatureProofs) Destroy() {
}
type InstructionCreateProofFromBucketOfAmount struct {
	BucketId ManifestBucket
	Amount *Decimal
}

func (e InstructionCreateProofFromBucketOfAmount) Destroy() {
		FfiDestroyerTypeManifestBucket{}.Destroy(e.BucketId);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type InstructionCreateProofFromBucketOfNonFungibles struct {
	BucketId ManifestBucket
	Ids []NonFungibleLocalId
}

func (e InstructionCreateProofFromBucketOfNonFungibles) Destroy() {
		FfiDestroyerTypeManifestBucket{}.Destroy(e.BucketId);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}
type InstructionCreateProofFromBucketOfAll struct {
	BucketId ManifestBucket
}

func (e InstructionCreateProofFromBucketOfAll) Destroy() {
		FfiDestroyerTypeManifestBucket{}.Destroy(e.BucketId);
}
type InstructionBurnResource struct {
	BucketId ManifestBucket
}

func (e InstructionBurnResource) Destroy() {
		FfiDestroyerTypeManifestBucket{}.Destroy(e.BucketId);
}
type InstructionCloneProof struct {
	ProofId ManifestProof
}

func (e InstructionCloneProof) Destroy() {
		FfiDestroyerTypeManifestProof{}.Destroy(e.ProofId);
}
type InstructionDropProof struct {
	ProofId ManifestProof
}

func (e InstructionDropProof) Destroy() {
		FfiDestroyerTypeManifestProof{}.Destroy(e.ProofId);
}
type InstructionCallFunction struct {
	PackageAddress ManifestAddress
	BlueprintName string
	FunctionName string
	Args ManifestValue
}

func (e InstructionCallFunction) Destroy() {
		FfiDestroyerTypeManifestAddress{}.Destroy(e.PackageAddress);
		FfiDestroyerString{}.Destroy(e.BlueprintName);
		FfiDestroyerString{}.Destroy(e.FunctionName);
		FfiDestroyerTypeManifestValue{}.Destroy(e.Args);
}
type InstructionCallMethod struct {
	Address ManifestAddress
	MethodName string
	Args ManifestValue
}

func (e InstructionCallMethod) Destroy() {
		FfiDestroyerTypeManifestAddress{}.Destroy(e.Address);
		FfiDestroyerString{}.Destroy(e.MethodName);
		FfiDestroyerTypeManifestValue{}.Destroy(e.Args);
}
type InstructionCallRoyaltyMethod struct {
	Address ManifestAddress
	MethodName string
	Args ManifestValue
}

func (e InstructionCallRoyaltyMethod) Destroy() {
		FfiDestroyerTypeManifestAddress{}.Destroy(e.Address);
		FfiDestroyerString{}.Destroy(e.MethodName);
		FfiDestroyerTypeManifestValue{}.Destroy(e.Args);
}
type InstructionCallMetadataMethod struct {
	Address ManifestAddress
	MethodName string
	Args ManifestValue
}

func (e InstructionCallMetadataMethod) Destroy() {
		FfiDestroyerTypeManifestAddress{}.Destroy(e.Address);
		FfiDestroyerString{}.Destroy(e.MethodName);
		FfiDestroyerTypeManifestValue{}.Destroy(e.Args);
}
type InstructionCallRoleAssignmentMethod struct {
	Address ManifestAddress
	MethodName string
	Args ManifestValue
}

func (e InstructionCallRoleAssignmentMethod) Destroy() {
		FfiDestroyerTypeManifestAddress{}.Destroy(e.Address);
		FfiDestroyerString{}.Destroy(e.MethodName);
		FfiDestroyerTypeManifestValue{}.Destroy(e.Args);
}
type InstructionCallDirectVaultMethod struct {
	Address *Address
	MethodName string
	Args ManifestValue
}

func (e InstructionCallDirectVaultMethod) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Address);
		FfiDestroyerString{}.Destroy(e.MethodName);
		FfiDestroyerTypeManifestValue{}.Destroy(e.Args);
}
type InstructionAllocateGlobalAddress struct {
	PackageAddress *Address
	BlueprintName string
}

func (e InstructionAllocateGlobalAddress) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.PackageAddress);
		FfiDestroyerString{}.Destroy(e.BlueprintName);
}

type FfiConverterTypeInstruction struct {}

var FfiConverterTypeInstructionINSTANCE = FfiConverterTypeInstruction{}

func (c FfiConverterTypeInstruction) Lift(rb RustBufferI) Instruction {
	return LiftFromRustBuffer[Instruction](c, rb)
}

func (c FfiConverterTypeInstruction) Lower(value Instruction) RustBuffer {
	return LowerIntoRustBuffer[Instruction](c, value)
}
func (FfiConverterTypeInstruction) Read(reader io.Reader) Instruction {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return InstructionTakeAllFromWorktop{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		case 2:
			return InstructionTakeFromWorktop{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 3:
			return InstructionTakeNonFungiblesFromWorktop{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 4:
			return InstructionReturnToWorktop{
				FfiConverterTypeManifestBucketINSTANCE.Read(reader),
			};
		case 5:
			return InstructionAssertWorktopContains{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 6:
			return InstructionAssertWorktopContainsAny{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		case 7:
			return InstructionAssertWorktopContainsNonFungibles{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 8:
			return InstructionPopFromAuthZone{
			};
		case 9:
			return InstructionPushToAuthZone{
				FfiConverterTypeManifestProofINSTANCE.Read(reader),
			};
		case 10:
			return InstructionCreateProofFromAuthZoneOfAmount{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 11:
			return InstructionCreateProofFromAuthZoneOfNonFungibles{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 12:
			return InstructionCreateProofFromAuthZoneOfAll{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		case 13:
			return InstructionDropAllProofs{
			};
		case 14:
			return InstructionDropNamedProofs{
			};
		case 15:
			return InstructionDropAuthZoneProofs{
			};
		case 16:
			return InstructionDropAuthZoneRegularProofs{
			};
		case 17:
			return InstructionDropAuthZoneSignatureProofs{
			};
		case 18:
			return InstructionCreateProofFromBucketOfAmount{
				FfiConverterTypeManifestBucketINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 19:
			return InstructionCreateProofFromBucketOfNonFungibles{
				FfiConverterTypeManifestBucketINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 20:
			return InstructionCreateProofFromBucketOfAll{
				FfiConverterTypeManifestBucketINSTANCE.Read(reader),
			};
		case 21:
			return InstructionBurnResource{
				FfiConverterTypeManifestBucketINSTANCE.Read(reader),
			};
		case 22:
			return InstructionCloneProof{
				FfiConverterTypeManifestProofINSTANCE.Read(reader),
			};
		case 23:
			return InstructionDropProof{
				FfiConverterTypeManifestProofINSTANCE.Read(reader),
			};
		case 24:
			return InstructionCallFunction{
				FfiConverterTypeManifestAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeManifestValueINSTANCE.Read(reader),
			};
		case 25:
			return InstructionCallMethod{
				FfiConverterTypeManifestAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeManifestValueINSTANCE.Read(reader),
			};
		case 26:
			return InstructionCallRoyaltyMethod{
				FfiConverterTypeManifestAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeManifestValueINSTANCE.Read(reader),
			};
		case 27:
			return InstructionCallMetadataMethod{
				FfiConverterTypeManifestAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeManifestValueINSTANCE.Read(reader),
			};
		case 28:
			return InstructionCallRoleAssignmentMethod{
				FfiConverterTypeManifestAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeManifestValueINSTANCE.Read(reader),
			};
		case 29:
			return InstructionCallDirectVaultMethod{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeManifestValueINSTANCE.Read(reader),
			};
		case 30:
			return InstructionAllocateGlobalAddress{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeInstruction.Read()", id));
	}
}

func (FfiConverterTypeInstruction) Write(writer io.Writer, value Instruction) {
	switch variant_value := value.(type) {
		case InstructionTakeAllFromWorktop:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
		case InstructionTakeFromWorktop:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case InstructionTakeNonFungiblesFromWorktop:
			writeInt32(writer, 3)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		case InstructionReturnToWorktop:
			writeInt32(writer, 4)
			FfiConverterTypeManifestBucketINSTANCE.Write(writer, variant_value.BucketId)
		case InstructionAssertWorktopContains:
			writeInt32(writer, 5)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case InstructionAssertWorktopContainsAny:
			writeInt32(writer, 6)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
		case InstructionAssertWorktopContainsNonFungibles:
			writeInt32(writer, 7)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		case InstructionPopFromAuthZone:
			writeInt32(writer, 8)
		case InstructionPushToAuthZone:
			writeInt32(writer, 9)
			FfiConverterTypeManifestProofINSTANCE.Write(writer, variant_value.ProofId)
		case InstructionCreateProofFromAuthZoneOfAmount:
			writeInt32(writer, 10)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case InstructionCreateProofFromAuthZoneOfNonFungibles:
			writeInt32(writer, 11)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		case InstructionCreateProofFromAuthZoneOfAll:
			writeInt32(writer, 12)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
		case InstructionDropAllProofs:
			writeInt32(writer, 13)
		case InstructionDropNamedProofs:
			writeInt32(writer, 14)
		case InstructionDropAuthZoneProofs:
			writeInt32(writer, 15)
		case InstructionDropAuthZoneRegularProofs:
			writeInt32(writer, 16)
		case InstructionDropAuthZoneSignatureProofs:
			writeInt32(writer, 17)
		case InstructionCreateProofFromBucketOfAmount:
			writeInt32(writer, 18)
			FfiConverterTypeManifestBucketINSTANCE.Write(writer, variant_value.BucketId)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case InstructionCreateProofFromBucketOfNonFungibles:
			writeInt32(writer, 19)
			FfiConverterTypeManifestBucketINSTANCE.Write(writer, variant_value.BucketId)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		case InstructionCreateProofFromBucketOfAll:
			writeInt32(writer, 20)
			FfiConverterTypeManifestBucketINSTANCE.Write(writer, variant_value.BucketId)
		case InstructionBurnResource:
			writeInt32(writer, 21)
			FfiConverterTypeManifestBucketINSTANCE.Write(writer, variant_value.BucketId)
		case InstructionCloneProof:
			writeInt32(writer, 22)
			FfiConverterTypeManifestProofINSTANCE.Write(writer, variant_value.ProofId)
		case InstructionDropProof:
			writeInt32(writer, 23)
			FfiConverterTypeManifestProofINSTANCE.Write(writer, variant_value.ProofId)
		case InstructionCallFunction:
			writeInt32(writer, 24)
			FfiConverterTypeManifestAddressINSTANCE.Write(writer, variant_value.PackageAddress)
			FfiConverterStringINSTANCE.Write(writer, variant_value.BlueprintName)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FunctionName)
			FfiConverterTypeManifestValueINSTANCE.Write(writer, variant_value.Args)
		case InstructionCallMethod:
			writeInt32(writer, 25)
			FfiConverterTypeManifestAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterStringINSTANCE.Write(writer, variant_value.MethodName)
			FfiConverterTypeManifestValueINSTANCE.Write(writer, variant_value.Args)
		case InstructionCallRoyaltyMethod:
			writeInt32(writer, 26)
			FfiConverterTypeManifestAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterStringINSTANCE.Write(writer, variant_value.MethodName)
			FfiConverterTypeManifestValueINSTANCE.Write(writer, variant_value.Args)
		case InstructionCallMetadataMethod:
			writeInt32(writer, 27)
			FfiConverterTypeManifestAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterStringINSTANCE.Write(writer, variant_value.MethodName)
			FfiConverterTypeManifestValueINSTANCE.Write(writer, variant_value.Args)
		case InstructionCallRoleAssignmentMethod:
			writeInt32(writer, 28)
			FfiConverterTypeManifestAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterStringINSTANCE.Write(writer, variant_value.MethodName)
			FfiConverterTypeManifestValueINSTANCE.Write(writer, variant_value.Args)
		case InstructionCallDirectVaultMethod:
			writeInt32(writer, 29)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Address)
			FfiConverterStringINSTANCE.Write(writer, variant_value.MethodName)
			FfiConverterTypeManifestValueINSTANCE.Write(writer, variant_value.Args)
		case InstructionAllocateGlobalAddress:
			writeInt32(writer, 30)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.PackageAddress)
			FfiConverterStringINSTANCE.Write(writer, variant_value.BlueprintName)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeInstruction.Write", value))
	}
}

type FfiDestroyerTypeInstruction struct {}

func (_ FfiDestroyerTypeInstruction) Destroy(value Instruction) {
	value.Destroy()
}




type LocalTypeId interface {
	Destroy()
}
type LocalTypeIdWellKnown struct {
	Value uint8
}

func (e LocalTypeIdWellKnown) Destroy() {
		FfiDestroyerUint8{}.Destroy(e.Value);
}
type LocalTypeIdSchemaLocalIndex struct {
	Value uint64
}

func (e LocalTypeIdSchemaLocalIndex) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.Value);
}

type FfiConverterTypeLocalTypeId struct {}

var FfiConverterTypeLocalTypeIdINSTANCE = FfiConverterTypeLocalTypeId{}

func (c FfiConverterTypeLocalTypeId) Lift(rb RustBufferI) LocalTypeId {
	return LiftFromRustBuffer[LocalTypeId](c, rb)
}

func (c FfiConverterTypeLocalTypeId) Lower(value LocalTypeId) RustBuffer {
	return LowerIntoRustBuffer[LocalTypeId](c, value)
}
func (FfiConverterTypeLocalTypeId) Read(reader io.Reader) LocalTypeId {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return LocalTypeIdWellKnown{
				FfiConverterUint8INSTANCE.Read(reader),
			};
		case 2:
			return LocalTypeIdSchemaLocalIndex{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeLocalTypeId.Read()", id));
	}
}

func (FfiConverterTypeLocalTypeId) Write(writer io.Writer, value LocalTypeId) {
	switch variant_value := value.(type) {
		case LocalTypeIdWellKnown:
			writeInt32(writer, 1)
			FfiConverterUint8INSTANCE.Write(writer, variant_value.Value)
		case LocalTypeIdSchemaLocalIndex:
			writeInt32(writer, 2)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeLocalTypeId.Write", value))
	}
}

type FfiDestroyerTypeLocalTypeId struct {}

func (_ FfiDestroyerTypeLocalTypeId) Destroy(value LocalTypeId) {
	value.Destroy()
}




type ManifestAddress interface {
	Destroy()
}
type ManifestAddressNamed struct {
	Value uint32
}

func (e ManifestAddressNamed) Destroy() {
		FfiDestroyerUint32{}.Destroy(e.Value);
}
type ManifestAddressStatic struct {
	Value *Address
}

func (e ManifestAddressStatic) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Value);
}

type FfiConverterTypeManifestAddress struct {}

var FfiConverterTypeManifestAddressINSTANCE = FfiConverterTypeManifestAddress{}

func (c FfiConverterTypeManifestAddress) Lift(rb RustBufferI) ManifestAddress {
	return LiftFromRustBuffer[ManifestAddress](c, rb)
}

func (c FfiConverterTypeManifestAddress) Lower(value ManifestAddress) RustBuffer {
	return LowerIntoRustBuffer[ManifestAddress](c, value)
}
func (FfiConverterTypeManifestAddress) Read(reader io.Reader) ManifestAddress {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ManifestAddressNamed{
				FfiConverterUint32INSTANCE.Read(reader),
			};
		case 2:
			return ManifestAddressStatic{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeManifestAddress.Read()", id));
	}
}

func (FfiConverterTypeManifestAddress) Write(writer io.Writer, value ManifestAddress) {
	switch variant_value := value.(type) {
		case ManifestAddressNamed:
			writeInt32(writer, 1)
			FfiConverterUint32INSTANCE.Write(writer, variant_value.Value)
		case ManifestAddressStatic:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeManifestAddress.Write", value))
	}
}

type FfiDestroyerTypeManifestAddress struct {}

func (_ FfiDestroyerTypeManifestAddress) Destroy(value ManifestAddress) {
	value.Destroy()
}




type ManifestBuilderAddress interface {
	Destroy()
}
type ManifestBuilderAddressNamed struct {
	Value ManifestBuilderNamedAddress
}

func (e ManifestBuilderAddressNamed) Destroy() {
		FfiDestroyerTypeManifestBuilderNamedAddress{}.Destroy(e.Value);
}
type ManifestBuilderAddressStatic struct {
	Value *Address
}

func (e ManifestBuilderAddressStatic) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Value);
}

type FfiConverterTypeManifestBuilderAddress struct {}

var FfiConverterTypeManifestBuilderAddressINSTANCE = FfiConverterTypeManifestBuilderAddress{}

func (c FfiConverterTypeManifestBuilderAddress) Lift(rb RustBufferI) ManifestBuilderAddress {
	return LiftFromRustBuffer[ManifestBuilderAddress](c, rb)
}

func (c FfiConverterTypeManifestBuilderAddress) Lower(value ManifestBuilderAddress) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderAddress](c, value)
}
func (FfiConverterTypeManifestBuilderAddress) Read(reader io.Reader) ManifestBuilderAddress {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ManifestBuilderAddressNamed{
				FfiConverterTypeManifestBuilderNamedAddressINSTANCE.Read(reader),
			};
		case 2:
			return ManifestBuilderAddressStatic{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeManifestBuilderAddress.Read()", id));
	}
}

func (FfiConverterTypeManifestBuilderAddress) Write(writer io.Writer, value ManifestBuilderAddress) {
	switch variant_value := value.(type) {
		case ManifestBuilderAddressNamed:
			writeInt32(writer, 1)
			FfiConverterTypeManifestBuilderNamedAddressINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderAddressStatic:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeManifestBuilderAddress.Write", value))
	}
}

type FfiDestroyerTypeManifestBuilderAddress struct {}

func (_ FfiDestroyerTypeManifestBuilderAddress) Destroy(value ManifestBuilderAddress) {
	value.Destroy()
}




type ManifestBuilderValue interface {
	Destroy()
}
type ManifestBuilderValueBoolValue struct {
	Value bool
}

func (e ManifestBuilderValueBoolValue) Destroy() {
		FfiDestroyerBool{}.Destroy(e.Value);
}
type ManifestBuilderValueI8Value struct {
	Value int8
}

func (e ManifestBuilderValueI8Value) Destroy() {
		FfiDestroyerInt8{}.Destroy(e.Value);
}
type ManifestBuilderValueI16Value struct {
	Value int16
}

func (e ManifestBuilderValueI16Value) Destroy() {
		FfiDestroyerInt16{}.Destroy(e.Value);
}
type ManifestBuilderValueI32Value struct {
	Value int32
}

func (e ManifestBuilderValueI32Value) Destroy() {
		FfiDestroyerInt32{}.Destroy(e.Value);
}
type ManifestBuilderValueI64Value struct {
	Value int64
}

func (e ManifestBuilderValueI64Value) Destroy() {
		FfiDestroyerInt64{}.Destroy(e.Value);
}
type ManifestBuilderValueI128Value struct {
	Value string
}

func (e ManifestBuilderValueI128Value) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type ManifestBuilderValueU8Value struct {
	Value uint8
}

func (e ManifestBuilderValueU8Value) Destroy() {
		FfiDestroyerUint8{}.Destroy(e.Value);
}
type ManifestBuilderValueU16Value struct {
	Value uint16
}

func (e ManifestBuilderValueU16Value) Destroy() {
		FfiDestroyerUint16{}.Destroy(e.Value);
}
type ManifestBuilderValueU32Value struct {
	Value uint32
}

func (e ManifestBuilderValueU32Value) Destroy() {
		FfiDestroyerUint32{}.Destroy(e.Value);
}
type ManifestBuilderValueU64Value struct {
	Value uint64
}

func (e ManifestBuilderValueU64Value) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.Value);
}
type ManifestBuilderValueU128Value struct {
	Value string
}

func (e ManifestBuilderValueU128Value) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type ManifestBuilderValueStringValue struct {
	Value string
}

func (e ManifestBuilderValueStringValue) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type ManifestBuilderValueEnumValue struct {
	Discriminator uint8
	Fields []ManifestBuilderValue
}

func (e ManifestBuilderValueEnumValue) Destroy() {
		FfiDestroyerUint8{}.Destroy(e.Discriminator);
		FfiDestroyerSequenceTypeManifestBuilderValue{}.Destroy(e.Fields);
}
type ManifestBuilderValueArrayValue struct {
	ElementValueKind ManifestBuilderValueKind
	Elements []ManifestBuilderValue
}

func (e ManifestBuilderValueArrayValue) Destroy() {
		FfiDestroyerTypeManifestBuilderValueKind{}.Destroy(e.ElementValueKind);
		FfiDestroyerSequenceTypeManifestBuilderValue{}.Destroy(e.Elements);
}
type ManifestBuilderValueTupleValue struct {
	Fields []ManifestBuilderValue
}

func (e ManifestBuilderValueTupleValue) Destroy() {
		FfiDestroyerSequenceTypeManifestBuilderValue{}.Destroy(e.Fields);
}
type ManifestBuilderValueMapValue struct {
	KeyValueKind ManifestBuilderValueKind
	ValueValueKind ManifestBuilderValueKind
	Entries []ManifestBuilderMapEntry
}

func (e ManifestBuilderValueMapValue) Destroy() {
		FfiDestroyerTypeManifestBuilderValueKind{}.Destroy(e.KeyValueKind);
		FfiDestroyerTypeManifestBuilderValueKind{}.Destroy(e.ValueValueKind);
		FfiDestroyerSequenceTypeManifestBuilderMapEntry{}.Destroy(e.Entries);
}
type ManifestBuilderValueAddressValue struct {
	Value ManifestBuilderAddress
}

func (e ManifestBuilderValueAddressValue) Destroy() {
		FfiDestroyerTypeManifestBuilderAddress{}.Destroy(e.Value);
}
type ManifestBuilderValueBucketValue struct {
	Value ManifestBuilderBucket
}

func (e ManifestBuilderValueBucketValue) Destroy() {
		FfiDestroyerTypeManifestBuilderBucket{}.Destroy(e.Value);
}
type ManifestBuilderValueProofValue struct {
	Value ManifestBuilderProof
}

func (e ManifestBuilderValueProofValue) Destroy() {
		FfiDestroyerTypeManifestBuilderProof{}.Destroy(e.Value);
}
type ManifestBuilderValueExpressionValue struct {
	Value ManifestExpression
}

func (e ManifestBuilderValueExpressionValue) Destroy() {
		FfiDestroyerTypeManifestExpression{}.Destroy(e.Value);
}
type ManifestBuilderValueBlobValue struct {
	Value ManifestBlobRef
}

func (e ManifestBuilderValueBlobValue) Destroy() {
		FfiDestroyerTypeManifestBlobRef{}.Destroy(e.Value);
}
type ManifestBuilderValueDecimalValue struct {
	Value *Decimal
}

func (e ManifestBuilderValueDecimalValue) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type ManifestBuilderValuePreciseDecimalValue struct {
	Value *PreciseDecimal
}

func (e ManifestBuilderValuePreciseDecimalValue) Destroy() {
		FfiDestroyerPreciseDecimal{}.Destroy(e.Value);
}
type ManifestBuilderValueNonFungibleLocalIdValue struct {
	Value NonFungibleLocalId
}

func (e ManifestBuilderValueNonFungibleLocalIdValue) Destroy() {
		FfiDestroyerTypeNonFungibleLocalId{}.Destroy(e.Value);
}
type ManifestBuilderValueAddressReservationValue struct {
	Value ManifestBuilderAddressReservation
}

func (e ManifestBuilderValueAddressReservationValue) Destroy() {
		FfiDestroyerTypeManifestBuilderAddressReservation{}.Destroy(e.Value);
}

type FfiConverterTypeManifestBuilderValue struct {}

var FfiConverterTypeManifestBuilderValueINSTANCE = FfiConverterTypeManifestBuilderValue{}

func (c FfiConverterTypeManifestBuilderValue) Lift(rb RustBufferI) ManifestBuilderValue {
	return LiftFromRustBuffer[ManifestBuilderValue](c, rb)
}

func (c FfiConverterTypeManifestBuilderValue) Lower(value ManifestBuilderValue) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderValue](c, value)
}
func (FfiConverterTypeManifestBuilderValue) Read(reader io.Reader) ManifestBuilderValue {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ManifestBuilderValueBoolValue{
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 2:
			return ManifestBuilderValueI8Value{
				FfiConverterInt8INSTANCE.Read(reader),
			};
		case 3:
			return ManifestBuilderValueI16Value{
				FfiConverterInt16INSTANCE.Read(reader),
			};
		case 4:
			return ManifestBuilderValueI32Value{
				FfiConverterInt32INSTANCE.Read(reader),
			};
		case 5:
			return ManifestBuilderValueI64Value{
				FfiConverterInt64INSTANCE.Read(reader),
			};
		case 6:
			return ManifestBuilderValueI128Value{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 7:
			return ManifestBuilderValueU8Value{
				FfiConverterUint8INSTANCE.Read(reader),
			};
		case 8:
			return ManifestBuilderValueU16Value{
				FfiConverterUint16INSTANCE.Read(reader),
			};
		case 9:
			return ManifestBuilderValueU32Value{
				FfiConverterUint32INSTANCE.Read(reader),
			};
		case 10:
			return ManifestBuilderValueU64Value{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 11:
			return ManifestBuilderValueU128Value{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 12:
			return ManifestBuilderValueStringValue{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 13:
			return ManifestBuilderValueEnumValue{
				FfiConverterUint8INSTANCE.Read(reader),
				FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Read(reader),
			};
		case 14:
			return ManifestBuilderValueArrayValue{
				FfiConverterTypeManifestBuilderValueKindINSTANCE.Read(reader),
				FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Read(reader),
			};
		case 15:
			return ManifestBuilderValueTupleValue{
				FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Read(reader),
			};
		case 16:
			return ManifestBuilderValueMapValue{
				FfiConverterTypeManifestBuilderValueKindINSTANCE.Read(reader),
				FfiConverterTypeManifestBuilderValueKindINSTANCE.Read(reader),
				FfiConverterSequenceTypeManifestBuilderMapEntryINSTANCE.Read(reader),
			};
		case 17:
			return ManifestBuilderValueAddressValue{
				FfiConverterTypeManifestBuilderAddressINSTANCE.Read(reader),
			};
		case 18:
			return ManifestBuilderValueBucketValue{
				FfiConverterTypeManifestBuilderBucketINSTANCE.Read(reader),
			};
		case 19:
			return ManifestBuilderValueProofValue{
				FfiConverterTypeManifestBuilderProofINSTANCE.Read(reader),
			};
		case 20:
			return ManifestBuilderValueExpressionValue{
				FfiConverterTypeManifestExpressionINSTANCE.Read(reader),
			};
		case 21:
			return ManifestBuilderValueBlobValue{
				FfiConverterTypeManifestBlobRefINSTANCE.Read(reader),
			};
		case 22:
			return ManifestBuilderValueDecimalValue{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 23:
			return ManifestBuilderValuePreciseDecimalValue{
				FfiConverterPreciseDecimalINSTANCE.Read(reader),
			};
		case 24:
			return ManifestBuilderValueNonFungibleLocalIdValue{
				FfiConverterTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 25:
			return ManifestBuilderValueAddressReservationValue{
				FfiConverterTypeManifestBuilderAddressReservationINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeManifestBuilderValue.Read()", id));
	}
}

func (FfiConverterTypeManifestBuilderValue) Write(writer io.Writer, value ManifestBuilderValue) {
	switch variant_value := value.(type) {
		case ManifestBuilderValueBoolValue:
			writeInt32(writer, 1)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueI8Value:
			writeInt32(writer, 2)
			FfiConverterInt8INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueI16Value:
			writeInt32(writer, 3)
			FfiConverterInt16INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueI32Value:
			writeInt32(writer, 4)
			FfiConverterInt32INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueI64Value:
			writeInt32(writer, 5)
			FfiConverterInt64INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueI128Value:
			writeInt32(writer, 6)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueU8Value:
			writeInt32(writer, 7)
			FfiConverterUint8INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueU16Value:
			writeInt32(writer, 8)
			FfiConverterUint16INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueU32Value:
			writeInt32(writer, 9)
			FfiConverterUint32INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueU64Value:
			writeInt32(writer, 10)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueU128Value:
			writeInt32(writer, 11)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueStringValue:
			writeInt32(writer, 12)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueEnumValue:
			writeInt32(writer, 13)
			FfiConverterUint8INSTANCE.Write(writer, variant_value.Discriminator)
			FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Write(writer, variant_value.Fields)
		case ManifestBuilderValueArrayValue:
			writeInt32(writer, 14)
			FfiConverterTypeManifestBuilderValueKindINSTANCE.Write(writer, variant_value.ElementValueKind)
			FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Write(writer, variant_value.Elements)
		case ManifestBuilderValueTupleValue:
			writeInt32(writer, 15)
			FfiConverterSequenceTypeManifestBuilderValueINSTANCE.Write(writer, variant_value.Fields)
		case ManifestBuilderValueMapValue:
			writeInt32(writer, 16)
			FfiConverterTypeManifestBuilderValueKindINSTANCE.Write(writer, variant_value.KeyValueKind)
			FfiConverterTypeManifestBuilderValueKindINSTANCE.Write(writer, variant_value.ValueValueKind)
			FfiConverterSequenceTypeManifestBuilderMapEntryINSTANCE.Write(writer, variant_value.Entries)
		case ManifestBuilderValueAddressValue:
			writeInt32(writer, 17)
			FfiConverterTypeManifestBuilderAddressINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueBucketValue:
			writeInt32(writer, 18)
			FfiConverterTypeManifestBuilderBucketINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueProofValue:
			writeInt32(writer, 19)
			FfiConverterTypeManifestBuilderProofINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueExpressionValue:
			writeInt32(writer, 20)
			FfiConverterTypeManifestExpressionINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueBlobValue:
			writeInt32(writer, 21)
			FfiConverterTypeManifestBlobRefINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueDecimalValue:
			writeInt32(writer, 22)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValuePreciseDecimalValue:
			writeInt32(writer, 23)
			FfiConverterPreciseDecimalINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueNonFungibleLocalIdValue:
			writeInt32(writer, 24)
			FfiConverterTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		case ManifestBuilderValueAddressReservationValue:
			writeInt32(writer, 25)
			FfiConverterTypeManifestBuilderAddressReservationINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeManifestBuilderValue.Write", value))
	}
}

type FfiDestroyerTypeManifestBuilderValue struct {}

func (_ FfiDestroyerTypeManifestBuilderValue) Destroy(value ManifestBuilderValue) {
	value.Destroy()
}




type ManifestBuilderValueKind uint

const (
	ManifestBuilderValueKindBoolValue ManifestBuilderValueKind = 1
	ManifestBuilderValueKindI8Value ManifestBuilderValueKind = 2
	ManifestBuilderValueKindI16Value ManifestBuilderValueKind = 3
	ManifestBuilderValueKindI32Value ManifestBuilderValueKind = 4
	ManifestBuilderValueKindI64Value ManifestBuilderValueKind = 5
	ManifestBuilderValueKindI128Value ManifestBuilderValueKind = 6
	ManifestBuilderValueKindU8Value ManifestBuilderValueKind = 7
	ManifestBuilderValueKindU16Value ManifestBuilderValueKind = 8
	ManifestBuilderValueKindU32Value ManifestBuilderValueKind = 9
	ManifestBuilderValueKindU64Value ManifestBuilderValueKind = 10
	ManifestBuilderValueKindU128Value ManifestBuilderValueKind = 11
	ManifestBuilderValueKindStringValue ManifestBuilderValueKind = 12
	ManifestBuilderValueKindEnumValue ManifestBuilderValueKind = 13
	ManifestBuilderValueKindArrayValue ManifestBuilderValueKind = 14
	ManifestBuilderValueKindTupleValue ManifestBuilderValueKind = 15
	ManifestBuilderValueKindMapValue ManifestBuilderValueKind = 16
	ManifestBuilderValueKindAddressValue ManifestBuilderValueKind = 17
	ManifestBuilderValueKindBucketValue ManifestBuilderValueKind = 18
	ManifestBuilderValueKindProofValue ManifestBuilderValueKind = 19
	ManifestBuilderValueKindExpressionValue ManifestBuilderValueKind = 20
	ManifestBuilderValueKindBlobValue ManifestBuilderValueKind = 21
	ManifestBuilderValueKindDecimalValue ManifestBuilderValueKind = 22
	ManifestBuilderValueKindPreciseDecimalValue ManifestBuilderValueKind = 23
	ManifestBuilderValueKindNonFungibleLocalIdValue ManifestBuilderValueKind = 24
	ManifestBuilderValueKindAddressReservationValue ManifestBuilderValueKind = 25
)

type FfiConverterTypeManifestBuilderValueKind struct {}

var FfiConverterTypeManifestBuilderValueKindINSTANCE = FfiConverterTypeManifestBuilderValueKind{}

func (c FfiConverterTypeManifestBuilderValueKind) Lift(rb RustBufferI) ManifestBuilderValueKind {
	return LiftFromRustBuffer[ManifestBuilderValueKind](c, rb)
}

func (c FfiConverterTypeManifestBuilderValueKind) Lower(value ManifestBuilderValueKind) RustBuffer {
	return LowerIntoRustBuffer[ManifestBuilderValueKind](c, value)
}
func (FfiConverterTypeManifestBuilderValueKind) Read(reader io.Reader) ManifestBuilderValueKind {
	id := readInt32(reader)
	return ManifestBuilderValueKind(id)
}

func (FfiConverterTypeManifestBuilderValueKind) Write(writer io.Writer, value ManifestBuilderValueKind) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeManifestBuilderValueKind struct {}

func (_ FfiDestroyerTypeManifestBuilderValueKind) Destroy(value ManifestBuilderValueKind) {
}




type ManifestClass uint

const (
	ManifestClassGeneral ManifestClass = 1
	ManifestClassTransfer ManifestClass = 2
	ManifestClassPoolContribution ManifestClass = 3
	ManifestClassPoolRedemption ManifestClass = 4
	ManifestClassValidatorStake ManifestClass = 5
	ManifestClassValidatorUnstake ManifestClass = 6
	ManifestClassValidatorClaim ManifestClass = 7
	ManifestClassAccountDepositSettingsUpdate ManifestClass = 8
)

type FfiConverterTypeManifestClass struct {}

var FfiConverterTypeManifestClassINSTANCE = FfiConverterTypeManifestClass{}

func (c FfiConverterTypeManifestClass) Lift(rb RustBufferI) ManifestClass {
	return LiftFromRustBuffer[ManifestClass](c, rb)
}

func (c FfiConverterTypeManifestClass) Lower(value ManifestClass) RustBuffer {
	return LowerIntoRustBuffer[ManifestClass](c, value)
}
func (FfiConverterTypeManifestClass) Read(reader io.Reader) ManifestClass {
	id := readInt32(reader)
	return ManifestClass(id)
}

func (FfiConverterTypeManifestClass) Write(writer io.Writer, value ManifestClass) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeManifestClass struct {}

func (_ FfiDestroyerTypeManifestClass) Destroy(value ManifestClass) {
}




type ManifestExpression uint

const (
	ManifestExpressionEntireWorktop ManifestExpression = 1
	ManifestExpressionEntireAuthZone ManifestExpression = 2
)

type FfiConverterTypeManifestExpression struct {}

var FfiConverterTypeManifestExpressionINSTANCE = FfiConverterTypeManifestExpression{}

func (c FfiConverterTypeManifestExpression) Lift(rb RustBufferI) ManifestExpression {
	return LiftFromRustBuffer[ManifestExpression](c, rb)
}

func (c FfiConverterTypeManifestExpression) Lower(value ManifestExpression) RustBuffer {
	return LowerIntoRustBuffer[ManifestExpression](c, value)
}
func (FfiConverterTypeManifestExpression) Read(reader io.Reader) ManifestExpression {
	id := readInt32(reader)
	return ManifestExpression(id)
}

func (FfiConverterTypeManifestExpression) Write(writer io.Writer, value ManifestExpression) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeManifestExpression struct {}

func (_ FfiDestroyerTypeManifestExpression) Destroy(value ManifestExpression) {
}




type ManifestSborStringRepresentation interface {
	Destroy()
}
type ManifestSborStringRepresentationManifestString struct {
}

func (e ManifestSborStringRepresentationManifestString) Destroy() {
}
type ManifestSborStringRepresentationJson struct {
	Value SerializationMode
}

func (e ManifestSborStringRepresentationJson) Destroy() {
		FfiDestroyerTypeSerializationMode{}.Destroy(e.Value);
}

type FfiConverterTypeManifestSborStringRepresentation struct {}

var FfiConverterTypeManifestSborStringRepresentationINSTANCE = FfiConverterTypeManifestSborStringRepresentation{}

func (c FfiConverterTypeManifestSborStringRepresentation) Lift(rb RustBufferI) ManifestSborStringRepresentation {
	return LiftFromRustBuffer[ManifestSborStringRepresentation](c, rb)
}

func (c FfiConverterTypeManifestSborStringRepresentation) Lower(value ManifestSborStringRepresentation) RustBuffer {
	return LowerIntoRustBuffer[ManifestSborStringRepresentation](c, value)
}
func (FfiConverterTypeManifestSborStringRepresentation) Read(reader io.Reader) ManifestSborStringRepresentation {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ManifestSborStringRepresentationManifestString{
			};
		case 2:
			return ManifestSborStringRepresentationJson{
				FfiConverterTypeSerializationModeINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeManifestSborStringRepresentation.Read()", id));
	}
}

func (FfiConverterTypeManifestSborStringRepresentation) Write(writer io.Writer, value ManifestSborStringRepresentation) {
	switch variant_value := value.(type) {
		case ManifestSborStringRepresentationManifestString:
			writeInt32(writer, 1)
		case ManifestSborStringRepresentationJson:
			writeInt32(writer, 2)
			FfiConverterTypeSerializationModeINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeManifestSborStringRepresentation.Write", value))
	}
}

type FfiDestroyerTypeManifestSborStringRepresentation struct {}

func (_ FfiDestroyerTypeManifestSborStringRepresentation) Destroy(value ManifestSborStringRepresentation) {
	value.Destroy()
}




type ManifestValue interface {
	Destroy()
}
type ManifestValueBoolValue struct {
	Value bool
}

func (e ManifestValueBoolValue) Destroy() {
		FfiDestroyerBool{}.Destroy(e.Value);
}
type ManifestValueI8Value struct {
	Value int8
}

func (e ManifestValueI8Value) Destroy() {
		FfiDestroyerInt8{}.Destroy(e.Value);
}
type ManifestValueI16Value struct {
	Value int16
}

func (e ManifestValueI16Value) Destroy() {
		FfiDestroyerInt16{}.Destroy(e.Value);
}
type ManifestValueI32Value struct {
	Value int32
}

func (e ManifestValueI32Value) Destroy() {
		FfiDestroyerInt32{}.Destroy(e.Value);
}
type ManifestValueI64Value struct {
	Value int64
}

func (e ManifestValueI64Value) Destroy() {
		FfiDestroyerInt64{}.Destroy(e.Value);
}
type ManifestValueI128Value struct {
	Value string
}

func (e ManifestValueI128Value) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type ManifestValueU8Value struct {
	Value uint8
}

func (e ManifestValueU8Value) Destroy() {
		FfiDestroyerUint8{}.Destroy(e.Value);
}
type ManifestValueU16Value struct {
	Value uint16
}

func (e ManifestValueU16Value) Destroy() {
		FfiDestroyerUint16{}.Destroy(e.Value);
}
type ManifestValueU32Value struct {
	Value uint32
}

func (e ManifestValueU32Value) Destroy() {
		FfiDestroyerUint32{}.Destroy(e.Value);
}
type ManifestValueU64Value struct {
	Value uint64
}

func (e ManifestValueU64Value) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.Value);
}
type ManifestValueU128Value struct {
	Value string
}

func (e ManifestValueU128Value) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type ManifestValueStringValue struct {
	Value string
}

func (e ManifestValueStringValue) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type ManifestValueEnumValue struct {
	Discriminator uint8
	Fields []ManifestValue
}

func (e ManifestValueEnumValue) Destroy() {
		FfiDestroyerUint8{}.Destroy(e.Discriminator);
		FfiDestroyerSequenceTypeManifestValue{}.Destroy(e.Fields);
}
type ManifestValueArrayValue struct {
	ElementValueKind ManifestValueKind
	Elements []ManifestValue
}

func (e ManifestValueArrayValue) Destroy() {
		FfiDestroyerTypeManifestValueKind{}.Destroy(e.ElementValueKind);
		FfiDestroyerSequenceTypeManifestValue{}.Destroy(e.Elements);
}
type ManifestValueTupleValue struct {
	Fields []ManifestValue
}

func (e ManifestValueTupleValue) Destroy() {
		FfiDestroyerSequenceTypeManifestValue{}.Destroy(e.Fields);
}
type ManifestValueMapValue struct {
	KeyValueKind ManifestValueKind
	ValueValueKind ManifestValueKind
	Entries []MapEntry
}

func (e ManifestValueMapValue) Destroy() {
		FfiDestroyerTypeManifestValueKind{}.Destroy(e.KeyValueKind);
		FfiDestroyerTypeManifestValueKind{}.Destroy(e.ValueValueKind);
		FfiDestroyerSequenceTypeMapEntry{}.Destroy(e.Entries);
}
type ManifestValueAddressValue struct {
	Value ManifestAddress
}

func (e ManifestValueAddressValue) Destroy() {
		FfiDestroyerTypeManifestAddress{}.Destroy(e.Value);
}
type ManifestValueBucketValue struct {
	Value ManifestBucket
}

func (e ManifestValueBucketValue) Destroy() {
		FfiDestroyerTypeManifestBucket{}.Destroy(e.Value);
}
type ManifestValueProofValue struct {
	Value ManifestProof
}

func (e ManifestValueProofValue) Destroy() {
		FfiDestroyerTypeManifestProof{}.Destroy(e.Value);
}
type ManifestValueExpressionValue struct {
	Value ManifestExpression
}

func (e ManifestValueExpressionValue) Destroy() {
		FfiDestroyerTypeManifestExpression{}.Destroy(e.Value);
}
type ManifestValueBlobValue struct {
	Value ManifestBlobRef
}

func (e ManifestValueBlobValue) Destroy() {
		FfiDestroyerTypeManifestBlobRef{}.Destroy(e.Value);
}
type ManifestValueDecimalValue struct {
	Value *Decimal
}

func (e ManifestValueDecimalValue) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type ManifestValuePreciseDecimalValue struct {
	Value *PreciseDecimal
}

func (e ManifestValuePreciseDecimalValue) Destroy() {
		FfiDestroyerPreciseDecimal{}.Destroy(e.Value);
}
type ManifestValueNonFungibleLocalIdValue struct {
	Value NonFungibleLocalId
}

func (e ManifestValueNonFungibleLocalIdValue) Destroy() {
		FfiDestroyerTypeNonFungibleLocalId{}.Destroy(e.Value);
}
type ManifestValueAddressReservationValue struct {
	Value ManifestAddressReservation
}

func (e ManifestValueAddressReservationValue) Destroy() {
		FfiDestroyerTypeManifestAddressReservation{}.Destroy(e.Value);
}

type FfiConverterTypeManifestValue struct {}

var FfiConverterTypeManifestValueINSTANCE = FfiConverterTypeManifestValue{}

func (c FfiConverterTypeManifestValue) Lift(rb RustBufferI) ManifestValue {
	return LiftFromRustBuffer[ManifestValue](c, rb)
}

func (c FfiConverterTypeManifestValue) Lower(value ManifestValue) RustBuffer {
	return LowerIntoRustBuffer[ManifestValue](c, value)
}
func (FfiConverterTypeManifestValue) Read(reader io.Reader) ManifestValue {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ManifestValueBoolValue{
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 2:
			return ManifestValueI8Value{
				FfiConverterInt8INSTANCE.Read(reader),
			};
		case 3:
			return ManifestValueI16Value{
				FfiConverterInt16INSTANCE.Read(reader),
			};
		case 4:
			return ManifestValueI32Value{
				FfiConverterInt32INSTANCE.Read(reader),
			};
		case 5:
			return ManifestValueI64Value{
				FfiConverterInt64INSTANCE.Read(reader),
			};
		case 6:
			return ManifestValueI128Value{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 7:
			return ManifestValueU8Value{
				FfiConverterUint8INSTANCE.Read(reader),
			};
		case 8:
			return ManifestValueU16Value{
				FfiConverterUint16INSTANCE.Read(reader),
			};
		case 9:
			return ManifestValueU32Value{
				FfiConverterUint32INSTANCE.Read(reader),
			};
		case 10:
			return ManifestValueU64Value{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 11:
			return ManifestValueU128Value{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 12:
			return ManifestValueStringValue{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 13:
			return ManifestValueEnumValue{
				FfiConverterUint8INSTANCE.Read(reader),
				FfiConverterSequenceTypeManifestValueINSTANCE.Read(reader),
			};
		case 14:
			return ManifestValueArrayValue{
				FfiConverterTypeManifestValueKindINSTANCE.Read(reader),
				FfiConverterSequenceTypeManifestValueINSTANCE.Read(reader),
			};
		case 15:
			return ManifestValueTupleValue{
				FfiConverterSequenceTypeManifestValueINSTANCE.Read(reader),
			};
		case 16:
			return ManifestValueMapValue{
				FfiConverterTypeManifestValueKindINSTANCE.Read(reader),
				FfiConverterTypeManifestValueKindINSTANCE.Read(reader),
				FfiConverterSequenceTypeMapEntryINSTANCE.Read(reader),
			};
		case 17:
			return ManifestValueAddressValue{
				FfiConverterTypeManifestAddressINSTANCE.Read(reader),
			};
		case 18:
			return ManifestValueBucketValue{
				FfiConverterTypeManifestBucketINSTANCE.Read(reader),
			};
		case 19:
			return ManifestValueProofValue{
				FfiConverterTypeManifestProofINSTANCE.Read(reader),
			};
		case 20:
			return ManifestValueExpressionValue{
				FfiConverterTypeManifestExpressionINSTANCE.Read(reader),
			};
		case 21:
			return ManifestValueBlobValue{
				FfiConverterTypeManifestBlobRefINSTANCE.Read(reader),
			};
		case 22:
			return ManifestValueDecimalValue{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 23:
			return ManifestValuePreciseDecimalValue{
				FfiConverterPreciseDecimalINSTANCE.Read(reader),
			};
		case 24:
			return ManifestValueNonFungibleLocalIdValue{
				FfiConverterTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 25:
			return ManifestValueAddressReservationValue{
				FfiConverterTypeManifestAddressReservationINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeManifestValue.Read()", id));
	}
}

func (FfiConverterTypeManifestValue) Write(writer io.Writer, value ManifestValue) {
	switch variant_value := value.(type) {
		case ManifestValueBoolValue:
			writeInt32(writer, 1)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueI8Value:
			writeInt32(writer, 2)
			FfiConverterInt8INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueI16Value:
			writeInt32(writer, 3)
			FfiConverterInt16INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueI32Value:
			writeInt32(writer, 4)
			FfiConverterInt32INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueI64Value:
			writeInt32(writer, 5)
			FfiConverterInt64INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueI128Value:
			writeInt32(writer, 6)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueU8Value:
			writeInt32(writer, 7)
			FfiConverterUint8INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueU16Value:
			writeInt32(writer, 8)
			FfiConverterUint16INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueU32Value:
			writeInt32(writer, 9)
			FfiConverterUint32INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueU64Value:
			writeInt32(writer, 10)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Value)
		case ManifestValueU128Value:
			writeInt32(writer, 11)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueStringValue:
			writeInt32(writer, 12)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueEnumValue:
			writeInt32(writer, 13)
			FfiConverterUint8INSTANCE.Write(writer, variant_value.Discriminator)
			FfiConverterSequenceTypeManifestValueINSTANCE.Write(writer, variant_value.Fields)
		case ManifestValueArrayValue:
			writeInt32(writer, 14)
			FfiConverterTypeManifestValueKindINSTANCE.Write(writer, variant_value.ElementValueKind)
			FfiConverterSequenceTypeManifestValueINSTANCE.Write(writer, variant_value.Elements)
		case ManifestValueTupleValue:
			writeInt32(writer, 15)
			FfiConverterSequenceTypeManifestValueINSTANCE.Write(writer, variant_value.Fields)
		case ManifestValueMapValue:
			writeInt32(writer, 16)
			FfiConverterTypeManifestValueKindINSTANCE.Write(writer, variant_value.KeyValueKind)
			FfiConverterTypeManifestValueKindINSTANCE.Write(writer, variant_value.ValueValueKind)
			FfiConverterSequenceTypeMapEntryINSTANCE.Write(writer, variant_value.Entries)
		case ManifestValueAddressValue:
			writeInt32(writer, 17)
			FfiConverterTypeManifestAddressINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueBucketValue:
			writeInt32(writer, 18)
			FfiConverterTypeManifestBucketINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueProofValue:
			writeInt32(writer, 19)
			FfiConverterTypeManifestProofINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueExpressionValue:
			writeInt32(writer, 20)
			FfiConverterTypeManifestExpressionINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueBlobValue:
			writeInt32(writer, 21)
			FfiConverterTypeManifestBlobRefINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueDecimalValue:
			writeInt32(writer, 22)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case ManifestValuePreciseDecimalValue:
			writeInt32(writer, 23)
			FfiConverterPreciseDecimalINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueNonFungibleLocalIdValue:
			writeInt32(writer, 24)
			FfiConverterTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		case ManifestValueAddressReservationValue:
			writeInt32(writer, 25)
			FfiConverterTypeManifestAddressReservationINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeManifestValue.Write", value))
	}
}

type FfiDestroyerTypeManifestValue struct {}

func (_ FfiDestroyerTypeManifestValue) Destroy(value ManifestValue) {
	value.Destroy()
}




type ManifestValueKind uint

const (
	ManifestValueKindBoolValue ManifestValueKind = 1
	ManifestValueKindI8Value ManifestValueKind = 2
	ManifestValueKindI16Value ManifestValueKind = 3
	ManifestValueKindI32Value ManifestValueKind = 4
	ManifestValueKindI64Value ManifestValueKind = 5
	ManifestValueKindI128Value ManifestValueKind = 6
	ManifestValueKindU8Value ManifestValueKind = 7
	ManifestValueKindU16Value ManifestValueKind = 8
	ManifestValueKindU32Value ManifestValueKind = 9
	ManifestValueKindU64Value ManifestValueKind = 10
	ManifestValueKindU128Value ManifestValueKind = 11
	ManifestValueKindStringValue ManifestValueKind = 12
	ManifestValueKindEnumValue ManifestValueKind = 13
	ManifestValueKindArrayValue ManifestValueKind = 14
	ManifestValueKindTupleValue ManifestValueKind = 15
	ManifestValueKindMapValue ManifestValueKind = 16
	ManifestValueKindAddressValue ManifestValueKind = 17
	ManifestValueKindBucketValue ManifestValueKind = 18
	ManifestValueKindProofValue ManifestValueKind = 19
	ManifestValueKindExpressionValue ManifestValueKind = 20
	ManifestValueKindBlobValue ManifestValueKind = 21
	ManifestValueKindDecimalValue ManifestValueKind = 22
	ManifestValueKindPreciseDecimalValue ManifestValueKind = 23
	ManifestValueKindNonFungibleLocalIdValue ManifestValueKind = 24
	ManifestValueKindAddressReservationValue ManifestValueKind = 25
)

type FfiConverterTypeManifestValueKind struct {}

var FfiConverterTypeManifestValueKindINSTANCE = FfiConverterTypeManifestValueKind{}

func (c FfiConverterTypeManifestValueKind) Lift(rb RustBufferI) ManifestValueKind {
	return LiftFromRustBuffer[ManifestValueKind](c, rb)
}

func (c FfiConverterTypeManifestValueKind) Lower(value ManifestValueKind) RustBuffer {
	return LowerIntoRustBuffer[ManifestValueKind](c, value)
}
func (FfiConverterTypeManifestValueKind) Read(reader io.Reader) ManifestValueKind {
	id := readInt32(reader)
	return ManifestValueKind(id)
}

func (FfiConverterTypeManifestValueKind) Write(writer io.Writer, value ManifestValueKind) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeManifestValueKind struct {}

func (_ FfiDestroyerTypeManifestValueKind) Destroy(value ManifestValueKind) {
}




type Message interface {
	Destroy()
}
type MessageNone struct {
}

func (e MessageNone) Destroy() {
}
type MessagePlainText struct {
	Value PlainTextMessage
}

func (e MessagePlainText) Destroy() {
		FfiDestroyerTypePlainTextMessage{}.Destroy(e.Value);
}
type MessageEncrypted struct {
	Value EncryptedMessage
}

func (e MessageEncrypted) Destroy() {
		FfiDestroyerTypeEncryptedMessage{}.Destroy(e.Value);
}

type FfiConverterTypeMessage struct {}

var FfiConverterTypeMessageINSTANCE = FfiConverterTypeMessage{}

func (c FfiConverterTypeMessage) Lift(rb RustBufferI) Message {
	return LiftFromRustBuffer[Message](c, rb)
}

func (c FfiConverterTypeMessage) Lower(value Message) RustBuffer {
	return LowerIntoRustBuffer[Message](c, value)
}
func (FfiConverterTypeMessage) Read(reader io.Reader) Message {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return MessageNone{
			};
		case 2:
			return MessagePlainText{
				FfiConverterTypePlainTextMessageINSTANCE.Read(reader),
			};
		case 3:
			return MessageEncrypted{
				FfiConverterTypeEncryptedMessageINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeMessage.Read()", id));
	}
}

func (FfiConverterTypeMessage) Write(writer io.Writer, value Message) {
	switch variant_value := value.(type) {
		case MessageNone:
			writeInt32(writer, 1)
		case MessagePlainText:
			writeInt32(writer, 2)
			FfiConverterTypePlainTextMessageINSTANCE.Write(writer, variant_value.Value)
		case MessageEncrypted:
			writeInt32(writer, 3)
			FfiConverterTypeEncryptedMessageINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeMessage.Write", value))
	}
}

type FfiDestroyerTypeMessage struct {}

func (_ FfiDestroyerTypeMessage) Destroy(value Message) {
	value.Destroy()
}




type MessageContent interface {
	Destroy()
}
type MessageContentStr struct {
	Value string
}

func (e MessageContentStr) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type MessageContentBytes struct {
	Value []byte
}

func (e MessageContentBytes) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}

type FfiConverterTypeMessageContent struct {}

var FfiConverterTypeMessageContentINSTANCE = FfiConverterTypeMessageContent{}

func (c FfiConverterTypeMessageContent) Lift(rb RustBufferI) MessageContent {
	return LiftFromRustBuffer[MessageContent](c, rb)
}

func (c FfiConverterTypeMessageContent) Lower(value MessageContent) RustBuffer {
	return LowerIntoRustBuffer[MessageContent](c, value)
}
func (FfiConverterTypeMessageContent) Read(reader io.Reader) MessageContent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return MessageContentStr{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return MessageContentBytes{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeMessageContent.Read()", id));
	}
}

func (FfiConverterTypeMessageContent) Write(writer io.Writer, value MessageContent) {
	switch variant_value := value.(type) {
		case MessageContentStr:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case MessageContentBytes:
			writeInt32(writer, 2)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeMessageContent.Write", value))
	}
}

type FfiDestroyerTypeMessageContent struct {}

func (_ FfiDestroyerTypeMessageContent) Destroy(value MessageContent) {
	value.Destroy()
}




type MetadataValue interface {
	Destroy()
}
type MetadataValueStringValue struct {
	Value string
}

func (e MetadataValueStringValue) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type MetadataValueBoolValue struct {
	Value bool
}

func (e MetadataValueBoolValue) Destroy() {
		FfiDestroyerBool{}.Destroy(e.Value);
}
type MetadataValueU8Value struct {
	Value uint8
}

func (e MetadataValueU8Value) Destroy() {
		FfiDestroyerUint8{}.Destroy(e.Value);
}
type MetadataValueU32Value struct {
	Value uint32
}

func (e MetadataValueU32Value) Destroy() {
		FfiDestroyerUint32{}.Destroy(e.Value);
}
type MetadataValueU64Value struct {
	Value uint64
}

func (e MetadataValueU64Value) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.Value);
}
type MetadataValueI32Value struct {
	Value int32
}

func (e MetadataValueI32Value) Destroy() {
		FfiDestroyerInt32{}.Destroy(e.Value);
}
type MetadataValueI64Value struct {
	Value int64
}

func (e MetadataValueI64Value) Destroy() {
		FfiDestroyerInt64{}.Destroy(e.Value);
}
type MetadataValueDecimalValue struct {
	Value *Decimal
}

func (e MetadataValueDecimalValue) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type MetadataValueGlobalAddressValue struct {
	Value *Address
}

func (e MetadataValueGlobalAddressValue) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Value);
}
type MetadataValuePublicKeyValue struct {
	Value PublicKey
}

func (e MetadataValuePublicKeyValue) Destroy() {
		FfiDestroyerTypePublicKey{}.Destroy(e.Value);
}
type MetadataValueNonFungibleGlobalIdValue struct {
	Value *NonFungibleGlobalId
}

func (e MetadataValueNonFungibleGlobalIdValue) Destroy() {
		FfiDestroyerNonFungibleGlobalId{}.Destroy(e.Value);
}
type MetadataValueNonFungibleLocalIdValue struct {
	Value NonFungibleLocalId
}

func (e MetadataValueNonFungibleLocalIdValue) Destroy() {
		FfiDestroyerTypeNonFungibleLocalId{}.Destroy(e.Value);
}
type MetadataValueInstantValue struct {
	Value int64
}

func (e MetadataValueInstantValue) Destroy() {
		FfiDestroyerInt64{}.Destroy(e.Value);
}
type MetadataValueUrlValue struct {
	Value string
}

func (e MetadataValueUrlValue) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type MetadataValueOriginValue struct {
	Value string
}

func (e MetadataValueOriginValue) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type MetadataValuePublicKeyHashValue struct {
	Value PublicKeyHash
}

func (e MetadataValuePublicKeyHashValue) Destroy() {
		FfiDestroyerTypePublicKeyHash{}.Destroy(e.Value);
}
type MetadataValueStringArrayValue struct {
	Value []string
}

func (e MetadataValueStringArrayValue) Destroy() {
		FfiDestroyerSequenceString{}.Destroy(e.Value);
}
type MetadataValueBoolArrayValue struct {
	Value []bool
}

func (e MetadataValueBoolArrayValue) Destroy() {
		FfiDestroyerSequenceBool{}.Destroy(e.Value);
}
type MetadataValueU8ArrayValue struct {
	Value []byte
}

func (e MetadataValueU8ArrayValue) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}
type MetadataValueU32ArrayValue struct {
	Value []uint32
}

func (e MetadataValueU32ArrayValue) Destroy() {
		FfiDestroyerSequenceUint32{}.Destroy(e.Value);
}
type MetadataValueU64ArrayValue struct {
	Value []uint64
}

func (e MetadataValueU64ArrayValue) Destroy() {
		FfiDestroyerSequenceUint64{}.Destroy(e.Value);
}
type MetadataValueI32ArrayValue struct {
	Value []int32
}

func (e MetadataValueI32ArrayValue) Destroy() {
		FfiDestroyerSequenceInt32{}.Destroy(e.Value);
}
type MetadataValueI64ArrayValue struct {
	Value []int64
}

func (e MetadataValueI64ArrayValue) Destroy() {
		FfiDestroyerSequenceInt64{}.Destroy(e.Value);
}
type MetadataValueDecimalArrayValue struct {
	Value []*Decimal
}

func (e MetadataValueDecimalArrayValue) Destroy() {
		FfiDestroyerSequenceDecimal{}.Destroy(e.Value);
}
type MetadataValueGlobalAddressArrayValue struct {
	Value []*Address
}

func (e MetadataValueGlobalAddressArrayValue) Destroy() {
		FfiDestroyerSequenceAddress{}.Destroy(e.Value);
}
type MetadataValuePublicKeyArrayValue struct {
	Value []PublicKey
}

func (e MetadataValuePublicKeyArrayValue) Destroy() {
		FfiDestroyerSequenceTypePublicKey{}.Destroy(e.Value);
}
type MetadataValueNonFungibleGlobalIdArrayValue struct {
	Value []*NonFungibleGlobalId
}

func (e MetadataValueNonFungibleGlobalIdArrayValue) Destroy() {
		FfiDestroyerSequenceNonFungibleGlobalId{}.Destroy(e.Value);
}
type MetadataValueNonFungibleLocalIdArrayValue struct {
	Value []NonFungibleLocalId
}

func (e MetadataValueNonFungibleLocalIdArrayValue) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Value);
}
type MetadataValueInstantArrayValue struct {
	Value []int64
}

func (e MetadataValueInstantArrayValue) Destroy() {
		FfiDestroyerSequenceInt64{}.Destroy(e.Value);
}
type MetadataValueUrlArrayValue struct {
	Value []string
}

func (e MetadataValueUrlArrayValue) Destroy() {
		FfiDestroyerSequenceString{}.Destroy(e.Value);
}
type MetadataValueOriginArrayValue struct {
	Value []string
}

func (e MetadataValueOriginArrayValue) Destroy() {
		FfiDestroyerSequenceString{}.Destroy(e.Value);
}
type MetadataValuePublicKeyHashArrayValue struct {
	Value []PublicKeyHash
}

func (e MetadataValuePublicKeyHashArrayValue) Destroy() {
		FfiDestroyerSequenceTypePublicKeyHash{}.Destroy(e.Value);
}

type FfiConverterTypeMetadataValue struct {}

var FfiConverterTypeMetadataValueINSTANCE = FfiConverterTypeMetadataValue{}

func (c FfiConverterTypeMetadataValue) Lift(rb RustBufferI) MetadataValue {
	return LiftFromRustBuffer[MetadataValue](c, rb)
}

func (c FfiConverterTypeMetadataValue) Lower(value MetadataValue) RustBuffer {
	return LowerIntoRustBuffer[MetadataValue](c, value)
}
func (FfiConverterTypeMetadataValue) Read(reader io.Reader) MetadataValue {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return MetadataValueStringValue{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return MetadataValueBoolValue{
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 3:
			return MetadataValueU8Value{
				FfiConverterUint8INSTANCE.Read(reader),
			};
		case 4:
			return MetadataValueU32Value{
				FfiConverterUint32INSTANCE.Read(reader),
			};
		case 5:
			return MetadataValueU64Value{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 6:
			return MetadataValueI32Value{
				FfiConverterInt32INSTANCE.Read(reader),
			};
		case 7:
			return MetadataValueI64Value{
				FfiConverterInt64INSTANCE.Read(reader),
			};
		case 8:
			return MetadataValueDecimalValue{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 9:
			return MetadataValueGlobalAddressValue{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		case 10:
			return MetadataValuePublicKeyValue{
				FfiConverterTypePublicKeyINSTANCE.Read(reader),
			};
		case 11:
			return MetadataValueNonFungibleGlobalIdValue{
				FfiConverterNonFungibleGlobalIdINSTANCE.Read(reader),
			};
		case 12:
			return MetadataValueNonFungibleLocalIdValue{
				FfiConverterTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 13:
			return MetadataValueInstantValue{
				FfiConverterInt64INSTANCE.Read(reader),
			};
		case 14:
			return MetadataValueUrlValue{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 15:
			return MetadataValueOriginValue{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 16:
			return MetadataValuePublicKeyHashValue{
				FfiConverterTypePublicKeyHashINSTANCE.Read(reader),
			};
		case 17:
			return MetadataValueStringArrayValue{
				FfiConverterSequenceStringINSTANCE.Read(reader),
			};
		case 18:
			return MetadataValueBoolArrayValue{
				FfiConverterSequenceBoolINSTANCE.Read(reader),
			};
		case 19:
			return MetadataValueU8ArrayValue{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		case 20:
			return MetadataValueU32ArrayValue{
				FfiConverterSequenceUint32INSTANCE.Read(reader),
			};
		case 21:
			return MetadataValueU64ArrayValue{
				FfiConverterSequenceUint64INSTANCE.Read(reader),
			};
		case 22:
			return MetadataValueI32ArrayValue{
				FfiConverterSequenceInt32INSTANCE.Read(reader),
			};
		case 23:
			return MetadataValueI64ArrayValue{
				FfiConverterSequenceInt64INSTANCE.Read(reader),
			};
		case 24:
			return MetadataValueDecimalArrayValue{
				FfiConverterSequenceDecimalINSTANCE.Read(reader),
			};
		case 25:
			return MetadataValueGlobalAddressArrayValue{
				FfiConverterSequenceAddressINSTANCE.Read(reader),
			};
		case 26:
			return MetadataValuePublicKeyArrayValue{
				FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
			};
		case 27:
			return MetadataValueNonFungibleGlobalIdArrayValue{
				FfiConverterSequenceNonFungibleGlobalIdINSTANCE.Read(reader),
			};
		case 28:
			return MetadataValueNonFungibleLocalIdArrayValue{
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		case 29:
			return MetadataValueInstantArrayValue{
				FfiConverterSequenceInt64INSTANCE.Read(reader),
			};
		case 30:
			return MetadataValueUrlArrayValue{
				FfiConverterSequenceStringINSTANCE.Read(reader),
			};
		case 31:
			return MetadataValueOriginArrayValue{
				FfiConverterSequenceStringINSTANCE.Read(reader),
			};
		case 32:
			return MetadataValuePublicKeyHashArrayValue{
				FfiConverterSequenceTypePublicKeyHashINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeMetadataValue.Read()", id));
	}
}

func (FfiConverterTypeMetadataValue) Write(writer io.Writer, value MetadataValue) {
	switch variant_value := value.(type) {
		case MetadataValueStringValue:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueBoolValue:
			writeInt32(writer, 2)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueU8Value:
			writeInt32(writer, 3)
			FfiConverterUint8INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueU32Value:
			writeInt32(writer, 4)
			FfiConverterUint32INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueU64Value:
			writeInt32(writer, 5)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueI32Value:
			writeInt32(writer, 6)
			FfiConverterInt32INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueI64Value:
			writeInt32(writer, 7)
			FfiConverterInt64INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueDecimalValue:
			writeInt32(writer, 8)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueGlobalAddressValue:
			writeInt32(writer, 9)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Value)
		case MetadataValuePublicKeyValue:
			writeInt32(writer, 10)
			FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueNonFungibleGlobalIdValue:
			writeInt32(writer, 11)
			FfiConverterNonFungibleGlobalIdINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueNonFungibleLocalIdValue:
			writeInt32(writer, 12)
			FfiConverterTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueInstantValue:
			writeInt32(writer, 13)
			FfiConverterInt64INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueUrlValue:
			writeInt32(writer, 14)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueOriginValue:
			writeInt32(writer, 15)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case MetadataValuePublicKeyHashValue:
			writeInt32(writer, 16)
			FfiConverterTypePublicKeyHashINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueStringArrayValue:
			writeInt32(writer, 17)
			FfiConverterSequenceStringINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueBoolArrayValue:
			writeInt32(writer, 18)
			FfiConverterSequenceBoolINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueU8ArrayValue:
			writeInt32(writer, 19)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueU32ArrayValue:
			writeInt32(writer, 20)
			FfiConverterSequenceUint32INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueU64ArrayValue:
			writeInt32(writer, 21)
			FfiConverterSequenceUint64INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueI32ArrayValue:
			writeInt32(writer, 22)
			FfiConverterSequenceInt32INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueI64ArrayValue:
			writeInt32(writer, 23)
			FfiConverterSequenceInt64INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueDecimalArrayValue:
			writeInt32(writer, 24)
			FfiConverterSequenceDecimalINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueGlobalAddressArrayValue:
			writeInt32(writer, 25)
			FfiConverterSequenceAddressINSTANCE.Write(writer, variant_value.Value)
		case MetadataValuePublicKeyArrayValue:
			writeInt32(writer, 26)
			FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueNonFungibleGlobalIdArrayValue:
			writeInt32(writer, 27)
			FfiConverterSequenceNonFungibleGlobalIdINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueNonFungibleLocalIdArrayValue:
			writeInt32(writer, 28)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueInstantArrayValue:
			writeInt32(writer, 29)
			FfiConverterSequenceInt64INSTANCE.Write(writer, variant_value.Value)
		case MetadataValueUrlArrayValue:
			writeInt32(writer, 30)
			FfiConverterSequenceStringINSTANCE.Write(writer, variant_value.Value)
		case MetadataValueOriginArrayValue:
			writeInt32(writer, 31)
			FfiConverterSequenceStringINSTANCE.Write(writer, variant_value.Value)
		case MetadataValuePublicKeyHashArrayValue:
			writeInt32(writer, 32)
			FfiConverterSequenceTypePublicKeyHashINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeMetadataValue.Write", value))
	}
}

type FfiDestroyerTypeMetadataValue struct {}

func (_ FfiDestroyerTypeMetadataValue) Destroy(value MetadataValue) {
	value.Destroy()
}




type ModuleId uint

const (
	ModuleIdMain ModuleId = 1
	ModuleIdMetadata ModuleId = 2
	ModuleIdRoyalty ModuleId = 3
	ModuleIdRoleAssignment ModuleId = 4
)

type FfiConverterTypeModuleId struct {}

var FfiConverterTypeModuleIdINSTANCE = FfiConverterTypeModuleId{}

func (c FfiConverterTypeModuleId) Lift(rb RustBufferI) ModuleId {
	return LiftFromRustBuffer[ModuleId](c, rb)
}

func (c FfiConverterTypeModuleId) Lower(value ModuleId) RustBuffer {
	return LowerIntoRustBuffer[ModuleId](c, value)
}
func (FfiConverterTypeModuleId) Read(reader io.Reader) ModuleId {
	id := readInt32(reader)
	return ModuleId(id)
}

func (FfiConverterTypeModuleId) Write(writer io.Writer, value ModuleId) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeModuleId struct {}

func (_ FfiDestroyerTypeModuleId) Destroy(value ModuleId) {
}




type NameRecordError interface {
	Destroy()
}
type NameRecordErrorObjectNameIsAlreadyTaken struct {
	Object string
	Name string
}

func (e NameRecordErrorObjectNameIsAlreadyTaken) Destroy() {
		FfiDestroyerString{}.Destroy(e.Object);
		FfiDestroyerString{}.Destroy(e.Name);
}
type NameRecordErrorObjectDoesNotExist struct {
	Object string
	Name string
}

func (e NameRecordErrorObjectDoesNotExist) Destroy() {
		FfiDestroyerString{}.Destroy(e.Object);
		FfiDestroyerString{}.Destroy(e.Name);
}

type FfiConverterTypeNameRecordError struct {}

var FfiConverterTypeNameRecordErrorINSTANCE = FfiConverterTypeNameRecordError{}

func (c FfiConverterTypeNameRecordError) Lift(rb RustBufferI) NameRecordError {
	return LiftFromRustBuffer[NameRecordError](c, rb)
}

func (c FfiConverterTypeNameRecordError) Lower(value NameRecordError) RustBuffer {
	return LowerIntoRustBuffer[NameRecordError](c, value)
}
func (FfiConverterTypeNameRecordError) Read(reader io.Reader) NameRecordError {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return NameRecordErrorObjectNameIsAlreadyTaken{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return NameRecordErrorObjectDoesNotExist{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeNameRecordError.Read()", id));
	}
}

func (FfiConverterTypeNameRecordError) Write(writer io.Writer, value NameRecordError) {
	switch variant_value := value.(type) {
		case NameRecordErrorObjectNameIsAlreadyTaken:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Object)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Name)
		case NameRecordErrorObjectDoesNotExist:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Object)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Name)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeNameRecordError.Write", value))
	}
}

type FfiDestroyerTypeNameRecordError struct {}

func (_ FfiDestroyerTypeNameRecordError) Destroy(value NameRecordError) {
	value.Destroy()
}




type NonFungibleLocalId interface {
	Destroy()
}
type NonFungibleLocalIdInteger struct {
	Value uint64
}

func (e NonFungibleLocalIdInteger) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.Value);
}
type NonFungibleLocalIdStr struct {
	Value string
}

func (e NonFungibleLocalIdStr) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}
type NonFungibleLocalIdBytes struct {
	Value []byte
}

func (e NonFungibleLocalIdBytes) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}
type NonFungibleLocalIdRuid struct {
	Value []byte
}

func (e NonFungibleLocalIdRuid) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}

type FfiConverterTypeNonFungibleLocalId struct {}

var FfiConverterTypeNonFungibleLocalIdINSTANCE = FfiConverterTypeNonFungibleLocalId{}

func (c FfiConverterTypeNonFungibleLocalId) Lift(rb RustBufferI) NonFungibleLocalId {
	return LiftFromRustBuffer[NonFungibleLocalId](c, rb)
}

func (c FfiConverterTypeNonFungibleLocalId) Lower(value NonFungibleLocalId) RustBuffer {
	return LowerIntoRustBuffer[NonFungibleLocalId](c, value)
}
func (FfiConverterTypeNonFungibleLocalId) Read(reader io.Reader) NonFungibleLocalId {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return NonFungibleLocalIdInteger{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 2:
			return NonFungibleLocalIdStr{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 3:
			return NonFungibleLocalIdBytes{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		case 4:
			return NonFungibleLocalIdRuid{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeNonFungibleLocalId.Read()", id));
	}
}

func (FfiConverterTypeNonFungibleLocalId) Write(writer io.Writer, value NonFungibleLocalId) {
	switch variant_value := value.(type) {
		case NonFungibleLocalIdInteger:
			writeInt32(writer, 1)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Value)
		case NonFungibleLocalIdStr:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		case NonFungibleLocalIdBytes:
			writeInt32(writer, 3)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		case NonFungibleLocalIdRuid:
			writeInt32(writer, 4)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeNonFungibleLocalId.Write", value))
	}
}

type FfiDestroyerTypeNonFungibleLocalId struct {}

func (_ FfiDestroyerTypeNonFungibleLocalId) Destroy(value NonFungibleLocalId) {
	value.Destroy()
}




type NonFungibleResourceIndicator interface {
	Destroy()
}
type NonFungibleResourceIndicatorByAll struct {
	PredictedAmount PredictedDecimal
	PredictedIds PredictedNonFungibleIds
}

func (e NonFungibleResourceIndicatorByAll) Destroy() {
		FfiDestroyerTypePredictedDecimal{}.Destroy(e.PredictedAmount);
		FfiDestroyerTypePredictedNonFungibleIds{}.Destroy(e.PredictedIds);
}
type NonFungibleResourceIndicatorByAmount struct {
	Amount *Decimal
	PredictedIds PredictedNonFungibleIds
}

func (e NonFungibleResourceIndicatorByAmount) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Amount);
		FfiDestroyerTypePredictedNonFungibleIds{}.Destroy(e.PredictedIds);
}
type NonFungibleResourceIndicatorByIds struct {
	Ids []NonFungibleLocalId
}

func (e NonFungibleResourceIndicatorByIds) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}

type FfiConverterTypeNonFungibleResourceIndicator struct {}

var FfiConverterTypeNonFungibleResourceIndicatorINSTANCE = FfiConverterTypeNonFungibleResourceIndicator{}

func (c FfiConverterTypeNonFungibleResourceIndicator) Lift(rb RustBufferI) NonFungibleResourceIndicator {
	return LiftFromRustBuffer[NonFungibleResourceIndicator](c, rb)
}

func (c FfiConverterTypeNonFungibleResourceIndicator) Lower(value NonFungibleResourceIndicator) RustBuffer {
	return LowerIntoRustBuffer[NonFungibleResourceIndicator](c, value)
}
func (FfiConverterTypeNonFungibleResourceIndicator) Read(reader io.Reader) NonFungibleResourceIndicator {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return NonFungibleResourceIndicatorByAll{
				FfiConverterTypePredictedDecimalINSTANCE.Read(reader),
				FfiConverterTypePredictedNonFungibleIdsINSTANCE.Read(reader),
			};
		case 2:
			return NonFungibleResourceIndicatorByAmount{
				FfiConverterDecimalINSTANCE.Read(reader),
				FfiConverterTypePredictedNonFungibleIdsINSTANCE.Read(reader),
			};
		case 3:
			return NonFungibleResourceIndicatorByIds{
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeNonFungibleResourceIndicator.Read()", id));
	}
}

func (FfiConverterTypeNonFungibleResourceIndicator) Write(writer io.Writer, value NonFungibleResourceIndicator) {
	switch variant_value := value.(type) {
		case NonFungibleResourceIndicatorByAll:
			writeInt32(writer, 1)
			FfiConverterTypePredictedDecimalINSTANCE.Write(writer, variant_value.PredictedAmount)
			FfiConverterTypePredictedNonFungibleIdsINSTANCE.Write(writer, variant_value.PredictedIds)
		case NonFungibleResourceIndicatorByAmount:
			writeInt32(writer, 2)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
			FfiConverterTypePredictedNonFungibleIdsINSTANCE.Write(writer, variant_value.PredictedIds)
		case NonFungibleResourceIndicatorByIds:
			writeInt32(writer, 3)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeNonFungibleResourceIndicator.Write", value))
	}
}

type FfiDestroyerTypeNonFungibleResourceIndicator struct {}

func (_ FfiDestroyerTypeNonFungibleResourceIndicator) Destroy(value NonFungibleResourceIndicator) {
	value.Destroy()
}




type OlympiaNetwork uint

const (
	OlympiaNetworkMainnet OlympiaNetwork = 1
	OlympiaNetworkStokenet OlympiaNetwork = 2
	OlympiaNetworkReleasenet OlympiaNetwork = 3
	OlympiaNetworkRcNet OlympiaNetwork = 4
	OlympiaNetworkMilestonenet OlympiaNetwork = 5
	OlympiaNetworkDevopsnet OlympiaNetwork = 6
	OlympiaNetworkSandpitnet OlympiaNetwork = 7
	OlympiaNetworkLocalnet OlympiaNetwork = 8
)

type FfiConverterTypeOlympiaNetwork struct {}

var FfiConverterTypeOlympiaNetworkINSTANCE = FfiConverterTypeOlympiaNetwork{}

func (c FfiConverterTypeOlympiaNetwork) Lift(rb RustBufferI) OlympiaNetwork {
	return LiftFromRustBuffer[OlympiaNetwork](c, rb)
}

func (c FfiConverterTypeOlympiaNetwork) Lower(value OlympiaNetwork) RustBuffer {
	return LowerIntoRustBuffer[OlympiaNetwork](c, value)
}
func (FfiConverterTypeOlympiaNetwork) Read(reader io.Reader) OlympiaNetwork {
	id := readInt32(reader)
	return OlympiaNetwork(id)
}

func (FfiConverterTypeOlympiaNetwork) Write(writer io.Writer, value OlympiaNetwork) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeOlympiaNetwork struct {}

func (_ FfiDestroyerTypeOlympiaNetwork) Destroy(value OlympiaNetwork) {
}




type Operation uint

const (
	OperationAdd Operation = 1
	OperationRemove Operation = 2
)

type FfiConverterTypeOperation struct {}

var FfiConverterTypeOperationINSTANCE = FfiConverterTypeOperation{}

func (c FfiConverterTypeOperation) Lift(rb RustBufferI) Operation {
	return LiftFromRustBuffer[Operation](c, rb)
}

func (c FfiConverterTypeOperation) Lower(value Operation) RustBuffer {
	return LowerIntoRustBuffer[Operation](c, value)
}
func (FfiConverterTypeOperation) Read(reader io.Reader) Operation {
	id := readInt32(reader)
	return Operation(id)
}

func (FfiConverterTypeOperation) Write(writer io.Writer, value Operation) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeOperation struct {}

func (_ FfiDestroyerTypeOperation) Destroy(value Operation) {
}




type OwnerRole interface {
	Destroy()
}
type OwnerRoleNone struct {
}

func (e OwnerRoleNone) Destroy() {
}
type OwnerRoleFixed struct {
	Value *AccessRule
}

func (e OwnerRoleFixed) Destroy() {
		FfiDestroyerAccessRule{}.Destroy(e.Value);
}
type OwnerRoleUpdatable struct {
	Value *AccessRule
}

func (e OwnerRoleUpdatable) Destroy() {
		FfiDestroyerAccessRule{}.Destroy(e.Value);
}

type FfiConverterTypeOwnerRole struct {}

var FfiConverterTypeOwnerRoleINSTANCE = FfiConverterTypeOwnerRole{}

func (c FfiConverterTypeOwnerRole) Lift(rb RustBufferI) OwnerRole {
	return LiftFromRustBuffer[OwnerRole](c, rb)
}

func (c FfiConverterTypeOwnerRole) Lower(value OwnerRole) RustBuffer {
	return LowerIntoRustBuffer[OwnerRole](c, value)
}
func (FfiConverterTypeOwnerRole) Read(reader io.Reader) OwnerRole {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return OwnerRoleNone{
			};
		case 2:
			return OwnerRoleFixed{
				FfiConverterAccessRuleINSTANCE.Read(reader),
			};
		case 3:
			return OwnerRoleUpdatable{
				FfiConverterAccessRuleINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeOwnerRole.Read()", id));
	}
}

func (FfiConverterTypeOwnerRole) Write(writer io.Writer, value OwnerRole) {
	switch variant_value := value.(type) {
		case OwnerRoleNone:
			writeInt32(writer, 1)
		case OwnerRoleFixed:
			writeInt32(writer, 2)
			FfiConverterAccessRuleINSTANCE.Write(writer, variant_value.Value)
		case OwnerRoleUpdatable:
			writeInt32(writer, 3)
			FfiConverterAccessRuleINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeOwnerRole.Write", value))
	}
}

type FfiDestroyerTypeOwnerRole struct {}

func (_ FfiDestroyerTypeOwnerRole) Destroy(value OwnerRole) {
	value.Destroy()
}




type Proposer uint

const (
	ProposerPrimary Proposer = 1
	ProposerRecovery Proposer = 2
)

type FfiConverterTypeProposer struct {}

var FfiConverterTypeProposerINSTANCE = FfiConverterTypeProposer{}

func (c FfiConverterTypeProposer) Lift(rb RustBufferI) Proposer {
	return LiftFromRustBuffer[Proposer](c, rb)
}

func (c FfiConverterTypeProposer) Lower(value Proposer) RustBuffer {
	return LowerIntoRustBuffer[Proposer](c, value)
}
func (FfiConverterTypeProposer) Read(reader io.Reader) Proposer {
	id := readInt32(reader)
	return Proposer(id)
}

func (FfiConverterTypeProposer) Write(writer io.Writer, value Proposer) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeProposer struct {}

func (_ FfiDestroyerTypeProposer) Destroy(value Proposer) {
}




type PublicKey interface {
	Destroy()
}
type PublicKeySecp256k1 struct {
	Value []byte
}

func (e PublicKeySecp256k1) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}
type PublicKeyEd25519 struct {
	Value []byte
}

func (e PublicKeyEd25519) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}

type FfiConverterTypePublicKey struct {}

var FfiConverterTypePublicKeyINSTANCE = FfiConverterTypePublicKey{}

func (c FfiConverterTypePublicKey) Lift(rb RustBufferI) PublicKey {
	return LiftFromRustBuffer[PublicKey](c, rb)
}

func (c FfiConverterTypePublicKey) Lower(value PublicKey) RustBuffer {
	return LowerIntoRustBuffer[PublicKey](c, value)
}
func (FfiConverterTypePublicKey) Read(reader io.Reader) PublicKey {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return PublicKeySecp256k1{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		case 2:
			return PublicKeyEd25519{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypePublicKey.Read()", id));
	}
}

func (FfiConverterTypePublicKey) Write(writer io.Writer, value PublicKey) {
	switch variant_value := value.(type) {
		case PublicKeySecp256k1:
			writeInt32(writer, 1)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		case PublicKeyEd25519:
			writeInt32(writer, 2)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypePublicKey.Write", value))
	}
}

type FfiDestroyerTypePublicKey struct {}

func (_ FfiDestroyerTypePublicKey) Destroy(value PublicKey) {
	value.Destroy()
}




type PublicKeyHash interface {
	Destroy()
}
type PublicKeyHashSecp256k1 struct {
	Value []byte
}

func (e PublicKeyHashSecp256k1) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}
type PublicKeyHashEd25519 struct {
	Value []byte
}

func (e PublicKeyHashEd25519) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}

type FfiConverterTypePublicKeyHash struct {}

var FfiConverterTypePublicKeyHashINSTANCE = FfiConverterTypePublicKeyHash{}

func (c FfiConverterTypePublicKeyHash) Lift(rb RustBufferI) PublicKeyHash {
	return LiftFromRustBuffer[PublicKeyHash](c, rb)
}

func (c FfiConverterTypePublicKeyHash) Lower(value PublicKeyHash) RustBuffer {
	return LowerIntoRustBuffer[PublicKeyHash](c, value)
}
func (FfiConverterTypePublicKeyHash) Read(reader io.Reader) PublicKeyHash {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return PublicKeyHashSecp256k1{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		case 2:
			return PublicKeyHashEd25519{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypePublicKeyHash.Read()", id));
	}
}

func (FfiConverterTypePublicKeyHash) Write(writer io.Writer, value PublicKeyHash) {
	switch variant_value := value.(type) {
		case PublicKeyHashSecp256k1:
			writeInt32(writer, 1)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		case PublicKeyHashEd25519:
			writeInt32(writer, 2)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypePublicKeyHash.Write", value))
	}
}

type FfiDestroyerTypePublicKeyHash struct {}

func (_ FfiDestroyerTypePublicKeyHash) Destroy(value PublicKeyHash) {
	value.Destroy()
}


type RadixEngineToolkitError struct {
	err error
}

func (err RadixEngineToolkitError) Error() string {
	return fmt.Sprintf("RadixEngineToolkitError: %s", err.err.Error())
}

func (err RadixEngineToolkitError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrRadixEngineToolkitErrorInvalidLength = fmt.Errorf("RadixEngineToolkitErrorInvalidLength")
var ErrRadixEngineToolkitErrorFailedToExtractNetwork = fmt.Errorf("RadixEngineToolkitErrorFailedToExtractNetwork")
var ErrRadixEngineToolkitErrorBech32DecodeError = fmt.Errorf("RadixEngineToolkitErrorBech32DecodeError")
var ErrRadixEngineToolkitErrorParseError = fmt.Errorf("RadixEngineToolkitErrorParseError")
var ErrRadixEngineToolkitErrorNonFungibleContentValidationError = fmt.Errorf("RadixEngineToolkitErrorNonFungibleContentValidationError")
var ErrRadixEngineToolkitErrorEntityTypeMismatchError = fmt.Errorf("RadixEngineToolkitErrorEntityTypeMismatchError")
var ErrRadixEngineToolkitErrorDerivationError = fmt.Errorf("RadixEngineToolkitErrorDerivationError")
var ErrRadixEngineToolkitErrorInvalidPublicKey = fmt.Errorf("RadixEngineToolkitErrorInvalidPublicKey")
var ErrRadixEngineToolkitErrorCompileError = fmt.Errorf("RadixEngineToolkitErrorCompileError")
var ErrRadixEngineToolkitErrorDecompileError = fmt.Errorf("RadixEngineToolkitErrorDecompileError")
var ErrRadixEngineToolkitErrorPrepareError = fmt.Errorf("RadixEngineToolkitErrorPrepareError")
var ErrRadixEngineToolkitErrorEncodeError = fmt.Errorf("RadixEngineToolkitErrorEncodeError")
var ErrRadixEngineToolkitErrorDecodeError = fmt.Errorf("RadixEngineToolkitErrorDecodeError")
var ErrRadixEngineToolkitErrorTransactionValidationFailed = fmt.Errorf("RadixEngineToolkitErrorTransactionValidationFailed")
var ErrRadixEngineToolkitErrorExecutionModuleError = fmt.Errorf("RadixEngineToolkitErrorExecutionModuleError")
var ErrRadixEngineToolkitErrorManifestSborError = fmt.Errorf("RadixEngineToolkitErrorManifestSborError")
var ErrRadixEngineToolkitErrorScryptoSborError = fmt.Errorf("RadixEngineToolkitErrorScryptoSborError")
var ErrRadixEngineToolkitErrorTypedNativeEventError = fmt.Errorf("RadixEngineToolkitErrorTypedNativeEventError")
var ErrRadixEngineToolkitErrorFailedToDecodeTransactionHash = fmt.Errorf("RadixEngineToolkitErrorFailedToDecodeTransactionHash")
var ErrRadixEngineToolkitErrorManifestBuilderNameRecordError = fmt.Errorf("RadixEngineToolkitErrorManifestBuilderNameRecordError")
var ErrRadixEngineToolkitErrorManifestModificationError = fmt.Errorf("RadixEngineToolkitErrorManifestModificationError")
var ErrRadixEngineToolkitErrorInvalidEntityTypeIdError = fmt.Errorf("RadixEngineToolkitErrorInvalidEntityTypeIdError")
var ErrRadixEngineToolkitErrorDecimalError = fmt.Errorf("RadixEngineToolkitErrorDecimalError")
var ErrRadixEngineToolkitErrorSignerError = fmt.Errorf("RadixEngineToolkitErrorSignerError")
var ErrRadixEngineToolkitErrorInvalidReceipt = fmt.Errorf("RadixEngineToolkitErrorInvalidReceipt")

// Variant structs
type RadixEngineToolkitErrorInvalidLength struct {
	Expected uint64
	Actual uint64
	Data []byte
}
func NewRadixEngineToolkitErrorInvalidLength(
	expected uint64,
	actual uint64,
	data []byte,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorInvalidLength{
			Expected: expected,
			Actual: actual,
			Data: data,
		},
	}
}

func (err RadixEngineToolkitErrorInvalidLength) Error() string {
	return fmt.Sprint("InvalidLength",
		": ",
		
		"Expected=",
		err.Expected,
		", ",
		"Actual=",
		err.Actual,
		", ",
		"Data=",
		err.Data,
	)
}

func (self RadixEngineToolkitErrorInvalidLength) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorInvalidLength
}
type RadixEngineToolkitErrorFailedToExtractNetwork struct {
	Address string
}
func NewRadixEngineToolkitErrorFailedToExtractNetwork(
	address string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorFailedToExtractNetwork{
			Address: address,
		},
	}
}

func (err RadixEngineToolkitErrorFailedToExtractNetwork) Error() string {
	return fmt.Sprint("FailedToExtractNetwork",
		": ",
		
		"Address=",
		err.Address,
	)
}

func (self RadixEngineToolkitErrorFailedToExtractNetwork) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorFailedToExtractNetwork
}
type RadixEngineToolkitErrorBech32DecodeError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorBech32DecodeError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorBech32DecodeError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorBech32DecodeError) Error() string {
	return fmt.Sprint("Bech32DecodeError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorBech32DecodeError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorBech32DecodeError
}
type RadixEngineToolkitErrorParseError struct {
	TypeName string
	Error_ string
}
func NewRadixEngineToolkitErrorParseError(
	typeName string,
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorParseError{
			TypeName: typeName,
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorParseError) Error() string {
	return fmt.Sprint("ParseError",
		": ",
		
		"TypeName=",
		err.TypeName,
		", ",
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorParseError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorParseError
}
type RadixEngineToolkitErrorNonFungibleContentValidationError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorNonFungibleContentValidationError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorNonFungibleContentValidationError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorNonFungibleContentValidationError) Error() string {
	return fmt.Sprint("NonFungibleContentValidationError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorNonFungibleContentValidationError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorNonFungibleContentValidationError
}
type RadixEngineToolkitErrorEntityTypeMismatchError struct {
	Expected []EntityType
	Actual EntityType
}
func NewRadixEngineToolkitErrorEntityTypeMismatchError(
	expected []EntityType,
	actual EntityType,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorEntityTypeMismatchError{
			Expected: expected,
			Actual: actual,
		},
	}
}

func (err RadixEngineToolkitErrorEntityTypeMismatchError) Error() string {
	return fmt.Sprint("EntityTypeMismatchError",
		": ",
		
		"Expected=",
		err.Expected,
		", ",
		"Actual=",
		err.Actual,
	)
}

func (self RadixEngineToolkitErrorEntityTypeMismatchError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorEntityTypeMismatchError
}
type RadixEngineToolkitErrorDerivationError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorDerivationError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorDerivationError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorDerivationError) Error() string {
	return fmt.Sprint("DerivationError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorDerivationError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorDerivationError
}
type RadixEngineToolkitErrorInvalidPublicKey struct {
}
func NewRadixEngineToolkitErrorInvalidPublicKey(
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorInvalidPublicKey{
		},
	}
}

func (err RadixEngineToolkitErrorInvalidPublicKey) Error() string {
	return fmt.Sprint("InvalidPublicKey",
		
	)
}

func (self RadixEngineToolkitErrorInvalidPublicKey) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorInvalidPublicKey
}
type RadixEngineToolkitErrorCompileError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorCompileError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorCompileError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorCompileError) Error() string {
	return fmt.Sprint("CompileError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorCompileError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorCompileError
}
type RadixEngineToolkitErrorDecompileError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorDecompileError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorDecompileError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorDecompileError) Error() string {
	return fmt.Sprint("DecompileError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorDecompileError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorDecompileError
}
type RadixEngineToolkitErrorPrepareError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorPrepareError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorPrepareError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorPrepareError) Error() string {
	return fmt.Sprint("PrepareError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorPrepareError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorPrepareError
}
type RadixEngineToolkitErrorEncodeError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorEncodeError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorEncodeError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorEncodeError) Error() string {
	return fmt.Sprint("EncodeError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorEncodeError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorEncodeError
}
type RadixEngineToolkitErrorDecodeError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorDecodeError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorDecodeError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorDecodeError) Error() string {
	return fmt.Sprint("DecodeError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorDecodeError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorDecodeError
}
type RadixEngineToolkitErrorTransactionValidationFailed struct {
	Error_ string
}
func NewRadixEngineToolkitErrorTransactionValidationFailed(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorTransactionValidationFailed{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorTransactionValidationFailed) Error() string {
	return fmt.Sprint("TransactionValidationFailed",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorTransactionValidationFailed) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorTransactionValidationFailed
}
type RadixEngineToolkitErrorExecutionModuleError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorExecutionModuleError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorExecutionModuleError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorExecutionModuleError) Error() string {
	return fmt.Sprint("ExecutionModuleError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorExecutionModuleError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorExecutionModuleError
}
type RadixEngineToolkitErrorManifestSborError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorManifestSborError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorManifestSborError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorManifestSborError) Error() string {
	return fmt.Sprint("ManifestSborError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorManifestSborError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorManifestSborError
}
type RadixEngineToolkitErrorScryptoSborError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorScryptoSborError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorScryptoSborError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorScryptoSborError) Error() string {
	return fmt.Sprint("ScryptoSborError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorScryptoSborError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorScryptoSborError
}
type RadixEngineToolkitErrorTypedNativeEventError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorTypedNativeEventError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorTypedNativeEventError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorTypedNativeEventError) Error() string {
	return fmt.Sprint("TypedNativeEventError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorTypedNativeEventError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorTypedNativeEventError
}
type RadixEngineToolkitErrorFailedToDecodeTransactionHash struct {
}
func NewRadixEngineToolkitErrorFailedToDecodeTransactionHash(
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorFailedToDecodeTransactionHash{
		},
	}
}

func (err RadixEngineToolkitErrorFailedToDecodeTransactionHash) Error() string {
	return fmt.Sprint("FailedToDecodeTransactionHash",
		
	)
}

func (self RadixEngineToolkitErrorFailedToDecodeTransactionHash) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorFailedToDecodeTransactionHash
}
type RadixEngineToolkitErrorManifestBuilderNameRecordError struct {
	Error_ NameRecordError
}
func NewRadixEngineToolkitErrorManifestBuilderNameRecordError(
	error NameRecordError,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorManifestBuilderNameRecordError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorManifestBuilderNameRecordError) Error() string {
	return fmt.Sprint("ManifestBuilderNameRecordError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorManifestBuilderNameRecordError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorManifestBuilderNameRecordError
}
type RadixEngineToolkitErrorManifestModificationError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorManifestModificationError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorManifestModificationError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorManifestModificationError) Error() string {
	return fmt.Sprint("ManifestModificationError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorManifestModificationError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorManifestModificationError
}
type RadixEngineToolkitErrorInvalidEntityTypeIdError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorInvalidEntityTypeIdError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorInvalidEntityTypeIdError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorInvalidEntityTypeIdError) Error() string {
	return fmt.Sprint("InvalidEntityTypeIdError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorInvalidEntityTypeIdError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorInvalidEntityTypeIdError
}
type RadixEngineToolkitErrorDecimalError struct {
}
func NewRadixEngineToolkitErrorDecimalError(
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorDecimalError{
		},
	}
}

func (err RadixEngineToolkitErrorDecimalError) Error() string {
	return fmt.Sprint("DecimalError",
		
	)
}

func (self RadixEngineToolkitErrorDecimalError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorDecimalError
}
type RadixEngineToolkitErrorSignerError struct {
	Error_ string
}
func NewRadixEngineToolkitErrorSignerError(
	error string,
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorSignerError{
			Error_: error,
		},
	}
}

func (err RadixEngineToolkitErrorSignerError) Error() string {
	return fmt.Sprint("SignerError",
		": ",
		
		"Error_=",
		err.Error_,
	)
}

func (self RadixEngineToolkitErrorSignerError) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorSignerError
}
type RadixEngineToolkitErrorInvalidReceipt struct {
}
func NewRadixEngineToolkitErrorInvalidReceipt(
) *RadixEngineToolkitError {
	return &RadixEngineToolkitError{
		err: &RadixEngineToolkitErrorInvalidReceipt{
		},
	}
}

func (err RadixEngineToolkitErrorInvalidReceipt) Error() string {
	return fmt.Sprint("InvalidReceipt",
		
	)
}

func (self RadixEngineToolkitErrorInvalidReceipt) Is(target error) bool {
	return target == ErrRadixEngineToolkitErrorInvalidReceipt
}

type FfiConverterTypeRadixEngineToolkitError struct{}

var FfiConverterTypeRadixEngineToolkitErrorINSTANCE = FfiConverterTypeRadixEngineToolkitError{}

func (c FfiConverterTypeRadixEngineToolkitError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeRadixEngineToolkitError) Lower(value *RadixEngineToolkitError) RustBuffer {
	return LowerIntoRustBuffer[*RadixEngineToolkitError](c, value)
}

func (c FfiConverterTypeRadixEngineToolkitError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorInvalidLength{
			Expected: FfiConverterUint64INSTANCE.Read(reader),
			Actual: FfiConverterUint64INSTANCE.Read(reader),
			Data: FfiConverterBytesINSTANCE.Read(reader),
		}}
	case 2:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorFailedToExtractNetwork{
			Address: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 3:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorBech32DecodeError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 4:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorParseError{
			TypeName: FfiConverterStringINSTANCE.Read(reader),
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 5:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorNonFungibleContentValidationError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 6:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorEntityTypeMismatchError{
			Expected: FfiConverterSequenceTypeEntityTypeINSTANCE.Read(reader),
			Actual: FfiConverterTypeEntityTypeINSTANCE.Read(reader),
		}}
	case 7:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorDerivationError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 8:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorInvalidPublicKey{
		}}
	case 9:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorCompileError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 10:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorDecompileError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 11:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorPrepareError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 12:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorEncodeError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 13:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorDecodeError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 14:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorTransactionValidationFailed{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 15:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorExecutionModuleError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 16:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorManifestSborError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 17:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorScryptoSborError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 18:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorTypedNativeEventError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 19:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorFailedToDecodeTransactionHash{
		}}
	case 20:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorManifestBuilderNameRecordError{
			Error_: FfiConverterTypeNameRecordErrorINSTANCE.Read(reader),
		}}
	case 21:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorManifestModificationError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 22:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorInvalidEntityTypeIdError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 23:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorDecimalError{
		}}
	case 24:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorSignerError{
			Error_: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 25:
		return &RadixEngineToolkitError{&RadixEngineToolkitErrorInvalidReceipt{
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeRadixEngineToolkitError.Read()", errorID))
	}
}

func (c FfiConverterTypeRadixEngineToolkitError) Write(writer io.Writer, value *RadixEngineToolkitError) {
	switch variantValue := value.err.(type) {
		case *RadixEngineToolkitErrorInvalidLength:
			writeInt32(writer, 1)
			FfiConverterUint64INSTANCE.Write(writer, variantValue.Expected)
			FfiConverterUint64INSTANCE.Write(writer, variantValue.Actual)
			FfiConverterBytesINSTANCE.Write(writer, variantValue.Data)
		case *RadixEngineToolkitErrorFailedToExtractNetwork:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Address)
		case *RadixEngineToolkitErrorBech32DecodeError:
			writeInt32(writer, 3)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorParseError:
			writeInt32(writer, 4)
			FfiConverterStringINSTANCE.Write(writer, variantValue.TypeName)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorNonFungibleContentValidationError:
			writeInt32(writer, 5)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorEntityTypeMismatchError:
			writeInt32(writer, 6)
			FfiConverterSequenceTypeEntityTypeINSTANCE.Write(writer, variantValue.Expected)
			FfiConverterTypeEntityTypeINSTANCE.Write(writer, variantValue.Actual)
		case *RadixEngineToolkitErrorDerivationError:
			writeInt32(writer, 7)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorInvalidPublicKey:
			writeInt32(writer, 8)
		case *RadixEngineToolkitErrorCompileError:
			writeInt32(writer, 9)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorDecompileError:
			writeInt32(writer, 10)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorPrepareError:
			writeInt32(writer, 11)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorEncodeError:
			writeInt32(writer, 12)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorDecodeError:
			writeInt32(writer, 13)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorTransactionValidationFailed:
			writeInt32(writer, 14)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorExecutionModuleError:
			writeInt32(writer, 15)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorManifestSborError:
			writeInt32(writer, 16)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorScryptoSborError:
			writeInt32(writer, 17)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorTypedNativeEventError:
			writeInt32(writer, 18)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorFailedToDecodeTransactionHash:
			writeInt32(writer, 19)
		case *RadixEngineToolkitErrorManifestBuilderNameRecordError:
			writeInt32(writer, 20)
			FfiConverterTypeNameRecordErrorINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorManifestModificationError:
			writeInt32(writer, 21)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorInvalidEntityTypeIdError:
			writeInt32(writer, 22)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorDecimalError:
			writeInt32(writer, 23)
		case *RadixEngineToolkitErrorSignerError:
			writeInt32(writer, 24)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Error_)
		case *RadixEngineToolkitErrorInvalidReceipt:
			writeInt32(writer, 25)
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeRadixEngineToolkitError.Write", value))
	}
}



type RecallResourceEvent interface {
	Destroy()
}
type RecallResourceEventAmount struct {
	Value *Decimal
}

func (e RecallResourceEventAmount) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type RecallResourceEventIds struct {
	Value []NonFungibleLocalId
}

func (e RecallResourceEventIds) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Value);
}

type FfiConverterTypeRecallResourceEvent struct {}

var FfiConverterTypeRecallResourceEventINSTANCE = FfiConverterTypeRecallResourceEvent{}

func (c FfiConverterTypeRecallResourceEvent) Lift(rb RustBufferI) RecallResourceEvent {
	return LiftFromRustBuffer[RecallResourceEvent](c, rb)
}

func (c FfiConverterTypeRecallResourceEvent) Lower(value RecallResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[RecallResourceEvent](c, value)
}
func (FfiConverterTypeRecallResourceEvent) Read(reader io.Reader) RecallResourceEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return RecallResourceEventAmount{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return RecallResourceEventIds{
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeRecallResourceEvent.Read()", id));
	}
}

func (FfiConverterTypeRecallResourceEvent) Write(writer io.Writer, value RecallResourceEvent) {
	switch variant_value := value.(type) {
		case RecallResourceEventAmount:
			writeInt32(writer, 1)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case RecallResourceEventIds:
			writeInt32(writer, 2)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeRecallResourceEvent.Write", value))
	}
}

type FfiDestroyerTypeRecallResourceEvent struct {}

func (_ FfiDestroyerTypeRecallResourceEvent) Destroy(value RecallResourceEvent) {
	value.Destroy()
}




type ReservedInstruction uint

const (
	ReservedInstructionAccountLockFee ReservedInstruction = 1
	ReservedInstructionAccountSecurify ReservedInstruction = 2
	ReservedInstructionIdentitySecurify ReservedInstruction = 3
	ReservedInstructionAccountUpdateSettings ReservedInstruction = 4
	ReservedInstructionAccessControllerMethod ReservedInstruction = 5
)

type FfiConverterTypeReservedInstruction struct {}

var FfiConverterTypeReservedInstructionINSTANCE = FfiConverterTypeReservedInstruction{}

func (c FfiConverterTypeReservedInstruction) Lift(rb RustBufferI) ReservedInstruction {
	return LiftFromRustBuffer[ReservedInstruction](c, rb)
}

func (c FfiConverterTypeReservedInstruction) Lower(value ReservedInstruction) RustBuffer {
	return LowerIntoRustBuffer[ReservedInstruction](c, value)
}
func (FfiConverterTypeReservedInstruction) Read(reader io.Reader) ReservedInstruction {
	id := readInt32(reader)
	return ReservedInstruction(id)
}

func (FfiConverterTypeReservedInstruction) Write(writer io.Writer, value ReservedInstruction) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeReservedInstruction struct {}

func (_ FfiDestroyerTypeReservedInstruction) Destroy(value ReservedInstruction) {
}




type ResourceIndicator interface {
	Destroy()
}
type ResourceIndicatorFungible struct {
	ResourceAddress *Address
	Indicator FungibleResourceIndicator
}

func (e ResourceIndicatorFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerTypeFungibleResourceIndicator{}.Destroy(e.Indicator);
}
type ResourceIndicatorNonFungible struct {
	ResourceAddress *Address
	Indicator NonFungibleResourceIndicator
}

func (e ResourceIndicatorNonFungible) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerTypeNonFungibleResourceIndicator{}.Destroy(e.Indicator);
}

type FfiConverterTypeResourceIndicator struct {}

var FfiConverterTypeResourceIndicatorINSTANCE = FfiConverterTypeResourceIndicator{}

func (c FfiConverterTypeResourceIndicator) Lift(rb RustBufferI) ResourceIndicator {
	return LiftFromRustBuffer[ResourceIndicator](c, rb)
}

func (c FfiConverterTypeResourceIndicator) Lower(value ResourceIndicator) RustBuffer {
	return LowerIntoRustBuffer[ResourceIndicator](c, value)
}
func (FfiConverterTypeResourceIndicator) Read(reader io.Reader) ResourceIndicator {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ResourceIndicatorFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterTypeFungibleResourceIndicatorINSTANCE.Read(reader),
			};
		case 2:
			return ResourceIndicatorNonFungible{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterTypeNonFungibleResourceIndicatorINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeResourceIndicator.Read()", id));
	}
}

func (FfiConverterTypeResourceIndicator) Write(writer io.Writer, value ResourceIndicator) {
	switch variant_value := value.(type) {
		case ResourceIndicatorFungible:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterTypeFungibleResourceIndicatorINSTANCE.Write(writer, variant_value.Indicator)
		case ResourceIndicatorNonFungible:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterTypeNonFungibleResourceIndicatorINSTANCE.Write(writer, variant_value.Indicator)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeResourceIndicator.Write", value))
	}
}

type FfiDestroyerTypeResourceIndicator struct {}

func (_ FfiDestroyerTypeResourceIndicator) Destroy(value ResourceIndicator) {
	value.Destroy()
}




type ResourceOrNonFungible interface {
	Destroy()
}
type ResourceOrNonFungibleNonFungible struct {
	Value *NonFungibleGlobalId
}

func (e ResourceOrNonFungibleNonFungible) Destroy() {
		FfiDestroyerNonFungibleGlobalId{}.Destroy(e.Value);
}
type ResourceOrNonFungibleResource struct {
	Value *Address
}

func (e ResourceOrNonFungibleResource) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.Value);
}

type FfiConverterTypeResourceOrNonFungible struct {}

var FfiConverterTypeResourceOrNonFungibleINSTANCE = FfiConverterTypeResourceOrNonFungible{}

func (c FfiConverterTypeResourceOrNonFungible) Lift(rb RustBufferI) ResourceOrNonFungible {
	return LiftFromRustBuffer[ResourceOrNonFungible](c, rb)
}

func (c FfiConverterTypeResourceOrNonFungible) Lower(value ResourceOrNonFungible) RustBuffer {
	return LowerIntoRustBuffer[ResourceOrNonFungible](c, value)
}
func (FfiConverterTypeResourceOrNonFungible) Read(reader io.Reader) ResourceOrNonFungible {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ResourceOrNonFungibleNonFungible{
				FfiConverterNonFungibleGlobalIdINSTANCE.Read(reader),
			};
		case 2:
			return ResourceOrNonFungibleResource{
				FfiConverterAddressINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeResourceOrNonFungible.Read()", id));
	}
}

func (FfiConverterTypeResourceOrNonFungible) Write(writer io.Writer, value ResourceOrNonFungible) {
	switch variant_value := value.(type) {
		case ResourceOrNonFungibleNonFungible:
			writeInt32(writer, 1)
			FfiConverterNonFungibleGlobalIdINSTANCE.Write(writer, variant_value.Value)
		case ResourceOrNonFungibleResource:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeResourceOrNonFungible.Write", value))
	}
}

type FfiDestroyerTypeResourceOrNonFungible struct {}

func (_ FfiDestroyerTypeResourceOrNonFungible) Destroy(value ResourceOrNonFungible) {
	value.Destroy()
}




type ResourcePreference uint

const (
	ResourcePreferenceAllowed ResourcePreference = 1
	ResourcePreferenceDisallowed ResourcePreference = 2
)

type FfiConverterTypeResourcePreference struct {}

var FfiConverterTypeResourcePreferenceINSTANCE = FfiConverterTypeResourcePreference{}

func (c FfiConverterTypeResourcePreference) Lift(rb RustBufferI) ResourcePreference {
	return LiftFromRustBuffer[ResourcePreference](c, rb)
}

func (c FfiConverterTypeResourcePreference) Lower(value ResourcePreference) RustBuffer {
	return LowerIntoRustBuffer[ResourcePreference](c, value)
}
func (FfiConverterTypeResourcePreference) Read(reader io.Reader) ResourcePreference {
	id := readInt32(reader)
	return ResourcePreference(id)
}

func (FfiConverterTypeResourcePreference) Write(writer io.Writer, value ResourcePreference) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeResourcePreference struct {}

func (_ FfiDestroyerTypeResourcePreference) Destroy(value ResourcePreference) {
}




type ResourcePreferenceUpdate interface {
	Destroy()
}
type ResourcePreferenceUpdateSet struct {
	Value ResourcePreference
}

func (e ResourcePreferenceUpdateSet) Destroy() {
		FfiDestroyerTypeResourcePreference{}.Destroy(e.Value);
}
type ResourcePreferenceUpdateRemove struct {
}

func (e ResourcePreferenceUpdateRemove) Destroy() {
}

type FfiConverterTypeResourcePreferenceUpdate struct {}

var FfiConverterTypeResourcePreferenceUpdateINSTANCE = FfiConverterTypeResourcePreferenceUpdate{}

func (c FfiConverterTypeResourcePreferenceUpdate) Lift(rb RustBufferI) ResourcePreferenceUpdate {
	return LiftFromRustBuffer[ResourcePreferenceUpdate](c, rb)
}

func (c FfiConverterTypeResourcePreferenceUpdate) Lower(value ResourcePreferenceUpdate) RustBuffer {
	return LowerIntoRustBuffer[ResourcePreferenceUpdate](c, value)
}
func (FfiConverterTypeResourcePreferenceUpdate) Read(reader io.Reader) ResourcePreferenceUpdate {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ResourcePreferenceUpdateSet{
				FfiConverterTypeResourcePreferenceINSTANCE.Read(reader),
			};
		case 2:
			return ResourcePreferenceUpdateRemove{
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeResourcePreferenceUpdate.Read()", id));
	}
}

func (FfiConverterTypeResourcePreferenceUpdate) Write(writer io.Writer, value ResourcePreferenceUpdate) {
	switch variant_value := value.(type) {
		case ResourcePreferenceUpdateSet:
			writeInt32(writer, 1)
			FfiConverterTypeResourcePreferenceINSTANCE.Write(writer, variant_value.Value)
		case ResourcePreferenceUpdateRemove:
			writeInt32(writer, 2)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeResourcePreferenceUpdate.Write", value))
	}
}

type FfiDestroyerTypeResourcePreferenceUpdate struct {}

func (_ FfiDestroyerTypeResourcePreferenceUpdate) Destroy(value ResourcePreferenceUpdate) {
	value.Destroy()
}




type ResourceSpecifier interface {
	Destroy()
}
type ResourceSpecifierAmount struct {
	ResourceAddress *Address
	Amount *Decimal
}

func (e ResourceSpecifierAmount) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerDecimal{}.Destroy(e.Amount);
}
type ResourceSpecifierIds struct {
	ResourceAddress *Address
	Ids []NonFungibleLocalId
}

func (e ResourceSpecifierIds) Destroy() {
		FfiDestroyerAddress{}.Destroy(e.ResourceAddress);
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Ids);
}

type FfiConverterTypeResourceSpecifier struct {}

var FfiConverterTypeResourceSpecifierINSTANCE = FfiConverterTypeResourceSpecifier{}

func (c FfiConverterTypeResourceSpecifier) Lift(rb RustBufferI) ResourceSpecifier {
	return LiftFromRustBuffer[ResourceSpecifier](c, rb)
}

func (c FfiConverterTypeResourceSpecifier) Lower(value ResourceSpecifier) RustBuffer {
	return LowerIntoRustBuffer[ResourceSpecifier](c, value)
}
func (FfiConverterTypeResourceSpecifier) Read(reader io.Reader) ResourceSpecifier {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ResourceSpecifierAmount{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return ResourceSpecifierIds{
				FfiConverterAddressINSTANCE.Read(reader),
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeResourceSpecifier.Read()", id));
	}
}

func (FfiConverterTypeResourceSpecifier) Write(writer io.Writer, value ResourceSpecifier) {
	switch variant_value := value.(type) {
		case ResourceSpecifierAmount:
			writeInt32(writer, 1)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Amount)
		case ResourceSpecifierIds:
			writeInt32(writer, 2)
			FfiConverterAddressINSTANCE.Write(writer, variant_value.ResourceAddress)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Ids)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeResourceSpecifier.Write", value))
	}
}

type FfiDestroyerTypeResourceSpecifier struct {}

func (_ FfiDestroyerTypeResourceSpecifier) Destroy(value ResourceSpecifier) {
	value.Destroy()
}




type Role uint

const (
	RolePrimary Role = 1
	RoleRecovery Role = 2
	RoleConfirmation Role = 3
)

type FfiConverterTypeRole struct {}

var FfiConverterTypeRoleINSTANCE = FfiConverterTypeRole{}

func (c FfiConverterTypeRole) Lift(rb RustBufferI) Role {
	return LiftFromRustBuffer[Role](c, rb)
}

func (c FfiConverterTypeRole) Lower(value Role) RustBuffer {
	return LowerIntoRustBuffer[Role](c, value)
}
func (FfiConverterTypeRole) Read(reader io.Reader) Role {
	id := readInt32(reader)
	return Role(id)
}

func (FfiConverterTypeRole) Write(writer io.Writer, value Role) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeRole struct {}

func (_ FfiDestroyerTypeRole) Destroy(value Role) {
}




type RoundingMode uint

const (
	RoundingModeToPositiveInfinity RoundingMode = 1
	RoundingModeToNegativeInfinity RoundingMode = 2
	RoundingModeToZero RoundingMode = 3
	RoundingModeAwayFromZero RoundingMode = 4
	RoundingModeToNearestMidpointTowardZero RoundingMode = 5
	RoundingModeToNearestMidpointAwayFromZero RoundingMode = 6
	RoundingModeToNearestMidpointToEven RoundingMode = 7
)

type FfiConverterTypeRoundingMode struct {}

var FfiConverterTypeRoundingModeINSTANCE = FfiConverterTypeRoundingMode{}

func (c FfiConverterTypeRoundingMode) Lift(rb RustBufferI) RoundingMode {
	return LiftFromRustBuffer[RoundingMode](c, rb)
}

func (c FfiConverterTypeRoundingMode) Lower(value RoundingMode) RustBuffer {
	return LowerIntoRustBuffer[RoundingMode](c, value)
}
func (FfiConverterTypeRoundingMode) Read(reader io.Reader) RoundingMode {
	id := readInt32(reader)
	return RoundingMode(id)
}

func (FfiConverterTypeRoundingMode) Write(writer io.Writer, value RoundingMode) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeRoundingMode struct {}

func (_ FfiDestroyerTypeRoundingMode) Destroy(value RoundingMode) {
}




type RoyaltyAmount interface {
	Destroy()
}
type RoyaltyAmountFree struct {
}

func (e RoyaltyAmountFree) Destroy() {
}
type RoyaltyAmountXrd struct {
	Value *Decimal
}

func (e RoyaltyAmountXrd) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type RoyaltyAmountUsd struct {
	Value *Decimal
}

func (e RoyaltyAmountUsd) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}

type FfiConverterTypeRoyaltyAmount struct {}

var FfiConverterTypeRoyaltyAmountINSTANCE = FfiConverterTypeRoyaltyAmount{}

func (c FfiConverterTypeRoyaltyAmount) Lift(rb RustBufferI) RoyaltyAmount {
	return LiftFromRustBuffer[RoyaltyAmount](c, rb)
}

func (c FfiConverterTypeRoyaltyAmount) Lower(value RoyaltyAmount) RustBuffer {
	return LowerIntoRustBuffer[RoyaltyAmount](c, value)
}
func (FfiConverterTypeRoyaltyAmount) Read(reader io.Reader) RoyaltyAmount {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return RoyaltyAmountFree{
			};
		case 2:
			return RoyaltyAmountXrd{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 3:
			return RoyaltyAmountUsd{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeRoyaltyAmount.Read()", id));
	}
}

func (FfiConverterTypeRoyaltyAmount) Write(writer io.Writer, value RoyaltyAmount) {
	switch variant_value := value.(type) {
		case RoyaltyAmountFree:
			writeInt32(writer, 1)
		case RoyaltyAmountXrd:
			writeInt32(writer, 2)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case RoyaltyAmountUsd:
			writeInt32(writer, 3)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeRoyaltyAmount.Write", value))
	}
}

type FfiDestroyerTypeRoyaltyAmount struct {}

func (_ FfiDestroyerTypeRoyaltyAmount) Destroy(value RoyaltyAmount) {
	value.Destroy()
}




type ScryptoSborString interface {
	Destroy()
}
type ScryptoSborStringProgrammaticJson struct {
	Value string
}

func (e ScryptoSborStringProgrammaticJson) Destroy() {
		FfiDestroyerString{}.Destroy(e.Value);
}

type FfiConverterTypeScryptoSborString struct {}

var FfiConverterTypeScryptoSborStringINSTANCE = FfiConverterTypeScryptoSborString{}

func (c FfiConverterTypeScryptoSborString) Lift(rb RustBufferI) ScryptoSborString {
	return LiftFromRustBuffer[ScryptoSborString](c, rb)
}

func (c FfiConverterTypeScryptoSborString) Lower(value ScryptoSborString) RustBuffer {
	return LowerIntoRustBuffer[ScryptoSborString](c, value)
}
func (FfiConverterTypeScryptoSborString) Read(reader io.Reader) ScryptoSborString {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return ScryptoSborStringProgrammaticJson{
				FfiConverterStringINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeScryptoSborString.Read()", id));
	}
}

func (FfiConverterTypeScryptoSborString) Write(writer io.Writer, value ScryptoSborString) {
	switch variant_value := value.(type) {
		case ScryptoSborStringProgrammaticJson:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeScryptoSborString.Write", value))
	}
}

type FfiDestroyerTypeScryptoSborString struct {}

func (_ FfiDestroyerTypeScryptoSborString) Destroy(value ScryptoSborString) {
	value.Destroy()
}




type SerializationMode uint

const (
	SerializationModeProgrammatic SerializationMode = 1
	SerializationModeNatural SerializationMode = 2
)

type FfiConverterTypeSerializationMode struct {}

var FfiConverterTypeSerializationModeINSTANCE = FfiConverterTypeSerializationMode{}

func (c FfiConverterTypeSerializationMode) Lift(rb RustBufferI) SerializationMode {
	return LiftFromRustBuffer[SerializationMode](c, rb)
}

func (c FfiConverterTypeSerializationMode) Lower(value SerializationMode) RustBuffer {
	return LowerIntoRustBuffer[SerializationMode](c, value)
}
func (FfiConverterTypeSerializationMode) Read(reader io.Reader) SerializationMode {
	id := readInt32(reader)
	return SerializationMode(id)
}

func (FfiConverterTypeSerializationMode) Write(writer io.Writer, value SerializationMode) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeSerializationMode struct {}

func (_ FfiDestroyerTypeSerializationMode) Destroy(value SerializationMode) {
}




type Signature interface {
	Destroy()
}
type SignatureSecp256k1 struct {
	Value []byte
}

func (e SignatureSecp256k1) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}
type SignatureEd25519 struct {
	Value []byte
}

func (e SignatureEd25519) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Value);
}

type FfiConverterTypeSignature struct {}

var FfiConverterTypeSignatureINSTANCE = FfiConverterTypeSignature{}

func (c FfiConverterTypeSignature) Lift(rb RustBufferI) Signature {
	return LiftFromRustBuffer[Signature](c, rb)
}

func (c FfiConverterTypeSignature) Lower(value Signature) RustBuffer {
	return LowerIntoRustBuffer[Signature](c, value)
}
func (FfiConverterTypeSignature) Read(reader io.Reader) Signature {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return SignatureSecp256k1{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		case 2:
			return SignatureEd25519{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeSignature.Read()", id));
	}
}

func (FfiConverterTypeSignature) Write(writer io.Writer, value Signature) {
	switch variant_value := value.(type) {
		case SignatureSecp256k1:
			writeInt32(writer, 1)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		case SignatureEd25519:
			writeInt32(writer, 2)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeSignature.Write", value))
	}
}

type FfiDestroyerTypeSignature struct {}

func (_ FfiDestroyerTypeSignature) Destroy(value Signature) {
	value.Destroy()
}




type SignatureWithPublicKey interface {
	Destroy()
}
type SignatureWithPublicKeySecp256k1 struct {
	Signature []byte
}

func (e SignatureWithPublicKeySecp256k1) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Signature);
}
type SignatureWithPublicKeyEd25519 struct {
	Signature []byte
	PublicKey []byte
}

func (e SignatureWithPublicKeyEd25519) Destroy() {
		FfiDestroyerBytes{}.Destroy(e.Signature);
		FfiDestroyerBytes{}.Destroy(e.PublicKey);
}

type FfiConverterTypeSignatureWithPublicKey struct {}

var FfiConverterTypeSignatureWithPublicKeyINSTANCE = FfiConverterTypeSignatureWithPublicKey{}

func (c FfiConverterTypeSignatureWithPublicKey) Lift(rb RustBufferI) SignatureWithPublicKey {
	return LiftFromRustBuffer[SignatureWithPublicKey](c, rb)
}

func (c FfiConverterTypeSignatureWithPublicKey) Lower(value SignatureWithPublicKey) RustBuffer {
	return LowerIntoRustBuffer[SignatureWithPublicKey](c, value)
}
func (FfiConverterTypeSignatureWithPublicKey) Read(reader io.Reader) SignatureWithPublicKey {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return SignatureWithPublicKeySecp256k1{
				FfiConverterBytesINSTANCE.Read(reader),
			};
		case 2:
			return SignatureWithPublicKeyEd25519{
				FfiConverterBytesINSTANCE.Read(reader),
				FfiConverterBytesINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeSignatureWithPublicKey.Read()", id));
	}
}

func (FfiConverterTypeSignatureWithPublicKey) Write(writer io.Writer, value SignatureWithPublicKey) {
	switch variant_value := value.(type) {
		case SignatureWithPublicKeySecp256k1:
			writeInt32(writer, 1)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Signature)
		case SignatureWithPublicKeyEd25519:
			writeInt32(writer, 2)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.Signature)
			FfiConverterBytesINSTANCE.Write(writer, variant_value.PublicKey)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeSignatureWithPublicKey.Write", value))
	}
}

type FfiDestroyerTypeSignatureWithPublicKey struct {}

func (_ FfiDestroyerTypeSignatureWithPublicKey) Destroy(value SignatureWithPublicKey) {
	value.Destroy()
}




type TypedAccessControllerBlueprintEvent interface {
	Destroy()
}
type TypedAccessControllerBlueprintEventInitiateRecoveryEventValue struct {
	Value InitiateRecoveryEvent
}

func (e TypedAccessControllerBlueprintEventInitiateRecoveryEventValue) Destroy() {
		FfiDestroyerTypeInitiateRecoveryEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventInitiateBadgeWithdrawAttemptEventValue struct {
	Value InitiateBadgeWithdrawAttemptEvent
}

func (e TypedAccessControllerBlueprintEventInitiateBadgeWithdrawAttemptEventValue) Destroy() {
		FfiDestroyerTypeInitiateBadgeWithdrawAttemptEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventRuleSetUpdateEventValue struct {
	Value RuleSetUpdateEvent
}

func (e TypedAccessControllerBlueprintEventRuleSetUpdateEventValue) Destroy() {
		FfiDestroyerTypeRuleSetUpdateEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventBadgeWithdrawEventValue struct {
	Value BadgeWithdrawEvent
}

func (e TypedAccessControllerBlueprintEventBadgeWithdrawEventValue) Destroy() {
		FfiDestroyerTypeBadgeWithdrawEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventCancelRecoveryProposalEventValue struct {
	Value CancelRecoveryProposalEvent
}

func (e TypedAccessControllerBlueprintEventCancelRecoveryProposalEventValue) Destroy() {
		FfiDestroyerTypeCancelRecoveryProposalEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventCancelBadgeWithdrawAttemptEventValue struct {
	Value CancelBadgeWithdrawAttemptEvent
}

func (e TypedAccessControllerBlueprintEventCancelBadgeWithdrawAttemptEventValue) Destroy() {
		FfiDestroyerTypeCancelBadgeWithdrawAttemptEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventLockPrimaryRoleEventValue struct {
	Value LockPrimaryRoleEvent
}

func (e TypedAccessControllerBlueprintEventLockPrimaryRoleEventValue) Destroy() {
		FfiDestroyerTypeLockPrimaryRoleEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventUnlockPrimaryRoleEventValue struct {
	Value UnlockPrimaryRoleEvent
}

func (e TypedAccessControllerBlueprintEventUnlockPrimaryRoleEventValue) Destroy() {
		FfiDestroyerTypeUnlockPrimaryRoleEvent{}.Destroy(e.Value);
}
type TypedAccessControllerBlueprintEventStopTimedRecoveryEventValue struct {
	Value StopTimedRecoveryEvent
}

func (e TypedAccessControllerBlueprintEventStopTimedRecoveryEventValue) Destroy() {
		FfiDestroyerTypeStopTimedRecoveryEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedAccessControllerBlueprintEvent struct {}

var FfiConverterTypeTypedAccessControllerBlueprintEventINSTANCE = FfiConverterTypeTypedAccessControllerBlueprintEvent{}

func (c FfiConverterTypeTypedAccessControllerBlueprintEvent) Lift(rb RustBufferI) TypedAccessControllerBlueprintEvent {
	return LiftFromRustBuffer[TypedAccessControllerBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedAccessControllerBlueprintEvent) Lower(value TypedAccessControllerBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedAccessControllerBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedAccessControllerBlueprintEvent) Read(reader io.Reader) TypedAccessControllerBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedAccessControllerBlueprintEventInitiateRecoveryEventValue{
				FfiConverterTypeInitiateRecoveryEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedAccessControllerBlueprintEventInitiateBadgeWithdrawAttemptEventValue{
				FfiConverterTypeInitiateBadgeWithdrawAttemptEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedAccessControllerBlueprintEventRuleSetUpdateEventValue{
				FfiConverterTypeRuleSetUpdateEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedAccessControllerBlueprintEventBadgeWithdrawEventValue{
				FfiConverterTypeBadgeWithdrawEventINSTANCE.Read(reader),
			};
		case 5:
			return TypedAccessControllerBlueprintEventCancelRecoveryProposalEventValue{
				FfiConverterTypeCancelRecoveryProposalEventINSTANCE.Read(reader),
			};
		case 6:
			return TypedAccessControllerBlueprintEventCancelBadgeWithdrawAttemptEventValue{
				FfiConverterTypeCancelBadgeWithdrawAttemptEventINSTANCE.Read(reader),
			};
		case 7:
			return TypedAccessControllerBlueprintEventLockPrimaryRoleEventValue{
				FfiConverterTypeLockPrimaryRoleEventINSTANCE.Read(reader),
			};
		case 8:
			return TypedAccessControllerBlueprintEventUnlockPrimaryRoleEventValue{
				FfiConverterTypeUnlockPrimaryRoleEventINSTANCE.Read(reader),
			};
		case 9:
			return TypedAccessControllerBlueprintEventStopTimedRecoveryEventValue{
				FfiConverterTypeStopTimedRecoveryEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedAccessControllerBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedAccessControllerBlueprintEvent) Write(writer io.Writer, value TypedAccessControllerBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedAccessControllerBlueprintEventInitiateRecoveryEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeInitiateRecoveryEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventInitiateBadgeWithdrawAttemptEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeInitiateBadgeWithdrawAttemptEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventRuleSetUpdateEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeRuleSetUpdateEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventBadgeWithdrawEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeBadgeWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventCancelRecoveryProposalEventValue:
			writeInt32(writer, 5)
			FfiConverterTypeCancelRecoveryProposalEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventCancelBadgeWithdrawAttemptEventValue:
			writeInt32(writer, 6)
			FfiConverterTypeCancelBadgeWithdrawAttemptEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventLockPrimaryRoleEventValue:
			writeInt32(writer, 7)
			FfiConverterTypeLockPrimaryRoleEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventUnlockPrimaryRoleEventValue:
			writeInt32(writer, 8)
			FfiConverterTypeUnlockPrimaryRoleEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccessControllerBlueprintEventStopTimedRecoveryEventValue:
			writeInt32(writer, 9)
			FfiConverterTypeStopTimedRecoveryEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedAccessControllerBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedAccessControllerBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedAccessControllerBlueprintEvent) Destroy(value TypedAccessControllerBlueprintEvent) {
	value.Destroy()
}




type TypedAccessControllerPackageEvent interface {
	Destroy()
}
type TypedAccessControllerPackageEventAccessController struct {
	Value TypedAccessControllerBlueprintEvent
}

func (e TypedAccessControllerPackageEventAccessController) Destroy() {
		FfiDestroyerTypeTypedAccessControllerBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedAccessControllerPackageEvent struct {}

var FfiConverterTypeTypedAccessControllerPackageEventINSTANCE = FfiConverterTypeTypedAccessControllerPackageEvent{}

func (c FfiConverterTypeTypedAccessControllerPackageEvent) Lift(rb RustBufferI) TypedAccessControllerPackageEvent {
	return LiftFromRustBuffer[TypedAccessControllerPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedAccessControllerPackageEvent) Lower(value TypedAccessControllerPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedAccessControllerPackageEvent](c, value)
}
func (FfiConverterTypeTypedAccessControllerPackageEvent) Read(reader io.Reader) TypedAccessControllerPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedAccessControllerPackageEventAccessController{
				FfiConverterTypeTypedAccessControllerBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedAccessControllerPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedAccessControllerPackageEvent) Write(writer io.Writer, value TypedAccessControllerPackageEvent) {
	switch variant_value := value.(type) {
		case TypedAccessControllerPackageEventAccessController:
			writeInt32(writer, 1)
			FfiConverterTypeTypedAccessControllerBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedAccessControllerPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedAccessControllerPackageEvent struct {}

func (_ FfiDestroyerTypeTypedAccessControllerPackageEvent) Destroy(value TypedAccessControllerPackageEvent) {
	value.Destroy()
}




type TypedAccountBlueprintEvent interface {
	Destroy()
}
type TypedAccountBlueprintEventAccountWithdrawEventValue struct {
	Value AccountWithdrawEvent
}

func (e TypedAccountBlueprintEventAccountWithdrawEventValue) Destroy() {
		FfiDestroyerTypeAccountWithdrawEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountDepositEventValue struct {
	Value AccountDepositEvent
}

func (e TypedAccountBlueprintEventAccountDepositEventValue) Destroy() {
		FfiDestroyerTypeAccountDepositEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountRejectedDepositEventValue struct {
	Value AccountRejectedDepositEvent
}

func (e TypedAccountBlueprintEventAccountRejectedDepositEventValue) Destroy() {
		FfiDestroyerTypeAccountRejectedDepositEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountSetResourcePreferenceEventValue struct {
	Value AccountSetResourcePreferenceEvent
}

func (e TypedAccountBlueprintEventAccountSetResourcePreferenceEventValue) Destroy() {
		FfiDestroyerTypeAccountSetResourcePreferenceEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountRemoveResourcePreferenceEventValue struct {
	Value AccountRemoveResourcePreferenceEvent
}

func (e TypedAccountBlueprintEventAccountRemoveResourcePreferenceEventValue) Destroy() {
		FfiDestroyerTypeAccountRemoveResourcePreferenceEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountSetDefaultDepositRuleEventValue struct {
	Value AccountSetDefaultDepositRuleEvent
}

func (e TypedAccountBlueprintEventAccountSetDefaultDepositRuleEventValue) Destroy() {
		FfiDestroyerTypeAccountSetDefaultDepositRuleEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountAddAuthorizedDepositorEventValue struct {
	Value AccountAddAuthorizedDepositorEvent
}

func (e TypedAccountBlueprintEventAccountAddAuthorizedDepositorEventValue) Destroy() {
		FfiDestroyerTypeAccountAddAuthorizedDepositorEvent{}.Destroy(e.Value);
}
type TypedAccountBlueprintEventAccountRemoveAuthorizedDepositorEventValue struct {
	Value AccountRemoveAuthorizedDepositorEvent
}

func (e TypedAccountBlueprintEventAccountRemoveAuthorizedDepositorEventValue) Destroy() {
		FfiDestroyerTypeAccountRemoveAuthorizedDepositorEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedAccountBlueprintEvent struct {}

var FfiConverterTypeTypedAccountBlueprintEventINSTANCE = FfiConverterTypeTypedAccountBlueprintEvent{}

func (c FfiConverterTypeTypedAccountBlueprintEvent) Lift(rb RustBufferI) TypedAccountBlueprintEvent {
	return LiftFromRustBuffer[TypedAccountBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedAccountBlueprintEvent) Lower(value TypedAccountBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedAccountBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedAccountBlueprintEvent) Read(reader io.Reader) TypedAccountBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedAccountBlueprintEventAccountWithdrawEventValue{
				FfiConverterTypeAccountWithdrawEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedAccountBlueprintEventAccountDepositEventValue{
				FfiConverterTypeAccountDepositEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedAccountBlueprintEventAccountRejectedDepositEventValue{
				FfiConverterTypeAccountRejectedDepositEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedAccountBlueprintEventAccountSetResourcePreferenceEventValue{
				FfiConverterTypeAccountSetResourcePreferenceEventINSTANCE.Read(reader),
			};
		case 5:
			return TypedAccountBlueprintEventAccountRemoveResourcePreferenceEventValue{
				FfiConverterTypeAccountRemoveResourcePreferenceEventINSTANCE.Read(reader),
			};
		case 6:
			return TypedAccountBlueprintEventAccountSetDefaultDepositRuleEventValue{
				FfiConverterTypeAccountSetDefaultDepositRuleEventINSTANCE.Read(reader),
			};
		case 7:
			return TypedAccountBlueprintEventAccountAddAuthorizedDepositorEventValue{
				FfiConverterTypeAccountAddAuthorizedDepositorEventINSTANCE.Read(reader),
			};
		case 8:
			return TypedAccountBlueprintEventAccountRemoveAuthorizedDepositorEventValue{
				FfiConverterTypeAccountRemoveAuthorizedDepositorEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedAccountBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedAccountBlueprintEvent) Write(writer io.Writer, value TypedAccountBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedAccountBlueprintEventAccountWithdrawEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeAccountWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountDepositEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeAccountDepositEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountRejectedDepositEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeAccountRejectedDepositEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountSetResourcePreferenceEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeAccountSetResourcePreferenceEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountRemoveResourcePreferenceEventValue:
			writeInt32(writer, 5)
			FfiConverterTypeAccountRemoveResourcePreferenceEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountSetDefaultDepositRuleEventValue:
			writeInt32(writer, 6)
			FfiConverterTypeAccountSetDefaultDepositRuleEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountAddAuthorizedDepositorEventValue:
			writeInt32(writer, 7)
			FfiConverterTypeAccountAddAuthorizedDepositorEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountBlueprintEventAccountRemoveAuthorizedDepositorEventValue:
			writeInt32(writer, 8)
			FfiConverterTypeAccountRemoveAuthorizedDepositorEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedAccountBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedAccountBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedAccountBlueprintEvent) Destroy(value TypedAccountBlueprintEvent) {
	value.Destroy()
}




type TypedAccountLockerBlueprintEvent interface {
	Destroy()
}
type TypedAccountLockerBlueprintEventStoreEventValue struct {
	Value StoreEvent
}

func (e TypedAccountLockerBlueprintEventStoreEventValue) Destroy() {
		FfiDestroyerTypeStoreEvent{}.Destroy(e.Value);
}
type TypedAccountLockerBlueprintEventRecoverEventValue struct {
	Value RecoverEvent
}

func (e TypedAccountLockerBlueprintEventRecoverEventValue) Destroy() {
		FfiDestroyerTypeRecoverEvent{}.Destroy(e.Value);
}
type TypedAccountLockerBlueprintEventClaimEventValue struct {
	Value ClaimEvent
}

func (e TypedAccountLockerBlueprintEventClaimEventValue) Destroy() {
		FfiDestroyerTypeClaimEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedAccountLockerBlueprintEvent struct {}

var FfiConverterTypeTypedAccountLockerBlueprintEventINSTANCE = FfiConverterTypeTypedAccountLockerBlueprintEvent{}

func (c FfiConverterTypeTypedAccountLockerBlueprintEvent) Lift(rb RustBufferI) TypedAccountLockerBlueprintEvent {
	return LiftFromRustBuffer[TypedAccountLockerBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedAccountLockerBlueprintEvent) Lower(value TypedAccountLockerBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedAccountLockerBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedAccountLockerBlueprintEvent) Read(reader io.Reader) TypedAccountLockerBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedAccountLockerBlueprintEventStoreEventValue{
				FfiConverterTypeStoreEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedAccountLockerBlueprintEventRecoverEventValue{
				FfiConverterTypeRecoverEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedAccountLockerBlueprintEventClaimEventValue{
				FfiConverterTypeClaimEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedAccountLockerBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedAccountLockerBlueprintEvent) Write(writer io.Writer, value TypedAccountLockerBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedAccountLockerBlueprintEventStoreEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeStoreEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountLockerBlueprintEventRecoverEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeRecoverEventINSTANCE.Write(writer, variant_value.Value)
		case TypedAccountLockerBlueprintEventClaimEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeClaimEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedAccountLockerBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedAccountLockerBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedAccountLockerBlueprintEvent) Destroy(value TypedAccountLockerBlueprintEvent) {
	value.Destroy()
}




type TypedAccountPackageEvent interface {
	Destroy()
}
type TypedAccountPackageEventAccount struct {
	Value TypedAccountBlueprintEvent
}

func (e TypedAccountPackageEventAccount) Destroy() {
		FfiDestroyerTypeTypedAccountBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedAccountPackageEvent struct {}

var FfiConverterTypeTypedAccountPackageEventINSTANCE = FfiConverterTypeTypedAccountPackageEvent{}

func (c FfiConverterTypeTypedAccountPackageEvent) Lift(rb RustBufferI) TypedAccountPackageEvent {
	return LiftFromRustBuffer[TypedAccountPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedAccountPackageEvent) Lower(value TypedAccountPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedAccountPackageEvent](c, value)
}
func (FfiConverterTypeTypedAccountPackageEvent) Read(reader io.Reader) TypedAccountPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedAccountPackageEventAccount{
				FfiConverterTypeTypedAccountBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedAccountPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedAccountPackageEvent) Write(writer io.Writer, value TypedAccountPackageEvent) {
	switch variant_value := value.(type) {
		case TypedAccountPackageEventAccount:
			writeInt32(writer, 1)
			FfiConverterTypeTypedAccountBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedAccountPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedAccountPackageEvent struct {}

func (_ FfiDestroyerTypeTypedAccountPackageEvent) Destroy(value TypedAccountPackageEvent) {
	value.Destroy()
}




type TypedConsensusManagerBlueprintEvent interface {
	Destroy()
}
type TypedConsensusManagerBlueprintEventRoundChangeEventValue struct {
	Value RoundChangeEvent
}

func (e TypedConsensusManagerBlueprintEventRoundChangeEventValue) Destroy() {
		FfiDestroyerTypeRoundChangeEvent{}.Destroy(e.Value);
}
type TypedConsensusManagerBlueprintEventEpochChangeEventValue struct {
	Value EpochChangeEvent
}

func (e TypedConsensusManagerBlueprintEventEpochChangeEventValue) Destroy() {
		FfiDestroyerTypeEpochChangeEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedConsensusManagerBlueprintEvent struct {}

var FfiConverterTypeTypedConsensusManagerBlueprintEventINSTANCE = FfiConverterTypeTypedConsensusManagerBlueprintEvent{}

func (c FfiConverterTypeTypedConsensusManagerBlueprintEvent) Lift(rb RustBufferI) TypedConsensusManagerBlueprintEvent {
	return LiftFromRustBuffer[TypedConsensusManagerBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedConsensusManagerBlueprintEvent) Lower(value TypedConsensusManagerBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedConsensusManagerBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedConsensusManagerBlueprintEvent) Read(reader io.Reader) TypedConsensusManagerBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedConsensusManagerBlueprintEventRoundChangeEventValue{
				FfiConverterTypeRoundChangeEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedConsensusManagerBlueprintEventEpochChangeEventValue{
				FfiConverterTypeEpochChangeEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedConsensusManagerBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedConsensusManagerBlueprintEvent) Write(writer io.Writer, value TypedConsensusManagerBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedConsensusManagerBlueprintEventRoundChangeEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeRoundChangeEventINSTANCE.Write(writer, variant_value.Value)
		case TypedConsensusManagerBlueprintEventEpochChangeEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeEpochChangeEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedConsensusManagerBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedConsensusManagerBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedConsensusManagerBlueprintEvent) Destroy(value TypedConsensusManagerBlueprintEvent) {
	value.Destroy()
}




type TypedConsensusManagerPackageEvent interface {
	Destroy()
}
type TypedConsensusManagerPackageEventConsensusManager struct {
	Value TypedConsensusManagerBlueprintEvent
}

func (e TypedConsensusManagerPackageEventConsensusManager) Destroy() {
		FfiDestroyerTypeTypedConsensusManagerBlueprintEvent{}.Destroy(e.Value);
}
type TypedConsensusManagerPackageEventValidator struct {
	Value TypedValidatorBlueprintEvent
}

func (e TypedConsensusManagerPackageEventValidator) Destroy() {
		FfiDestroyerTypeTypedValidatorBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedConsensusManagerPackageEvent struct {}

var FfiConverterTypeTypedConsensusManagerPackageEventINSTANCE = FfiConverterTypeTypedConsensusManagerPackageEvent{}

func (c FfiConverterTypeTypedConsensusManagerPackageEvent) Lift(rb RustBufferI) TypedConsensusManagerPackageEvent {
	return LiftFromRustBuffer[TypedConsensusManagerPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedConsensusManagerPackageEvent) Lower(value TypedConsensusManagerPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedConsensusManagerPackageEvent](c, value)
}
func (FfiConverterTypeTypedConsensusManagerPackageEvent) Read(reader io.Reader) TypedConsensusManagerPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedConsensusManagerPackageEventConsensusManager{
				FfiConverterTypeTypedConsensusManagerBlueprintEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedConsensusManagerPackageEventValidator{
				FfiConverterTypeTypedValidatorBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedConsensusManagerPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedConsensusManagerPackageEvent) Write(writer io.Writer, value TypedConsensusManagerPackageEvent) {
	switch variant_value := value.(type) {
		case TypedConsensusManagerPackageEventConsensusManager:
			writeInt32(writer, 1)
			FfiConverterTypeTypedConsensusManagerBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		case TypedConsensusManagerPackageEventValidator:
			writeInt32(writer, 2)
			FfiConverterTypeTypedValidatorBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedConsensusManagerPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedConsensusManagerPackageEvent struct {}

func (_ FfiDestroyerTypeTypedConsensusManagerPackageEvent) Destroy(value TypedConsensusManagerPackageEvent) {
	value.Destroy()
}




type TypedFungibleResourceManagerBlueprintEvent interface {
	Destroy()
}
type TypedFungibleResourceManagerBlueprintEventVaultCreationEventValue struct {
	Value VaultCreationEvent
}

func (e TypedFungibleResourceManagerBlueprintEventVaultCreationEventValue) Destroy() {
		FfiDestroyerTypeVaultCreationEvent{}.Destroy(e.Value);
}
type TypedFungibleResourceManagerBlueprintEventMintFungibleResourceEventValue struct {
	Value MintFungibleResourceEvent
}

func (e TypedFungibleResourceManagerBlueprintEventMintFungibleResourceEventValue) Destroy() {
		FfiDestroyerTypeMintFungibleResourceEvent{}.Destroy(e.Value);
}
type TypedFungibleResourceManagerBlueprintEventBurnFungibleResourceEventValue struct {
	Value BurnFungibleResourceEvent
}

func (e TypedFungibleResourceManagerBlueprintEventBurnFungibleResourceEventValue) Destroy() {
		FfiDestroyerTypeBurnFungibleResourceEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent struct {}

var FfiConverterTypeTypedFungibleResourceManagerBlueprintEventINSTANCE = FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent{}

func (c FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent) Lift(rb RustBufferI) TypedFungibleResourceManagerBlueprintEvent {
	return LiftFromRustBuffer[TypedFungibleResourceManagerBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent) Lower(value TypedFungibleResourceManagerBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedFungibleResourceManagerBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent) Read(reader io.Reader) TypedFungibleResourceManagerBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedFungibleResourceManagerBlueprintEventVaultCreationEventValue{
				FfiConverterTypeVaultCreationEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedFungibleResourceManagerBlueprintEventMintFungibleResourceEventValue{
				FfiConverterTypeMintFungibleResourceEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedFungibleResourceManagerBlueprintEventBurnFungibleResourceEventValue{
				FfiConverterTypeBurnFungibleResourceEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent) Write(writer io.Writer, value TypedFungibleResourceManagerBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedFungibleResourceManagerBlueprintEventVaultCreationEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeVaultCreationEventINSTANCE.Write(writer, variant_value.Value)
		case TypedFungibleResourceManagerBlueprintEventMintFungibleResourceEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeMintFungibleResourceEventINSTANCE.Write(writer, variant_value.Value)
		case TypedFungibleResourceManagerBlueprintEventBurnFungibleResourceEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeBurnFungibleResourceEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedFungibleResourceManagerBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedFungibleResourceManagerBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedFungibleResourceManagerBlueprintEvent) Destroy(value TypedFungibleResourceManagerBlueprintEvent) {
	value.Destroy()
}




type TypedFungibleVaultBlueprintEvent interface {
	Destroy()
}
type TypedFungibleVaultBlueprintEventFungibleVaultLockFeeEventValue struct {
	Value FungibleVaultLockFeeEvent
}

func (e TypedFungibleVaultBlueprintEventFungibleVaultLockFeeEventValue) Destroy() {
		FfiDestroyerTypeFungibleVaultLockFeeEvent{}.Destroy(e.Value);
}
type TypedFungibleVaultBlueprintEventFungibleVaultWithdrawEventValue struct {
	Value FungibleVaultWithdrawEvent
}

func (e TypedFungibleVaultBlueprintEventFungibleVaultWithdrawEventValue) Destroy() {
		FfiDestroyerTypeFungibleVaultWithdrawEvent{}.Destroy(e.Value);
}
type TypedFungibleVaultBlueprintEventFungibleVaultDepositEventValue struct {
	Value FungibleVaultDepositEvent
}

func (e TypedFungibleVaultBlueprintEventFungibleVaultDepositEventValue) Destroy() {
		FfiDestroyerTypeFungibleVaultDepositEvent{}.Destroy(e.Value);
}
type TypedFungibleVaultBlueprintEventFungibleVaultRecallEventValue struct {
	Value FungibleVaultRecallEvent
}

func (e TypedFungibleVaultBlueprintEventFungibleVaultRecallEventValue) Destroy() {
		FfiDestroyerTypeFungibleVaultRecallEvent{}.Destroy(e.Value);
}
type TypedFungibleVaultBlueprintEventFungibleVaultPayFeeEventValue struct {
	Value FungibleVaultPayFeeEvent
}

func (e TypedFungibleVaultBlueprintEventFungibleVaultPayFeeEventValue) Destroy() {
		FfiDestroyerTypeFungibleVaultPayFeeEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedFungibleVaultBlueprintEvent struct {}

var FfiConverterTypeTypedFungibleVaultBlueprintEventINSTANCE = FfiConverterTypeTypedFungibleVaultBlueprintEvent{}

func (c FfiConverterTypeTypedFungibleVaultBlueprintEvent) Lift(rb RustBufferI) TypedFungibleVaultBlueprintEvent {
	return LiftFromRustBuffer[TypedFungibleVaultBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedFungibleVaultBlueprintEvent) Lower(value TypedFungibleVaultBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedFungibleVaultBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedFungibleVaultBlueprintEvent) Read(reader io.Reader) TypedFungibleVaultBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedFungibleVaultBlueprintEventFungibleVaultLockFeeEventValue{
				FfiConverterTypeFungibleVaultLockFeeEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedFungibleVaultBlueprintEventFungibleVaultWithdrawEventValue{
				FfiConverterTypeFungibleVaultWithdrawEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedFungibleVaultBlueprintEventFungibleVaultDepositEventValue{
				FfiConverterTypeFungibleVaultDepositEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedFungibleVaultBlueprintEventFungibleVaultRecallEventValue{
				FfiConverterTypeFungibleVaultRecallEventINSTANCE.Read(reader),
			};
		case 5:
			return TypedFungibleVaultBlueprintEventFungibleVaultPayFeeEventValue{
				FfiConverterTypeFungibleVaultPayFeeEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedFungibleVaultBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedFungibleVaultBlueprintEvent) Write(writer io.Writer, value TypedFungibleVaultBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedFungibleVaultBlueprintEventFungibleVaultLockFeeEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeFungibleVaultLockFeeEventINSTANCE.Write(writer, variant_value.Value)
		case TypedFungibleVaultBlueprintEventFungibleVaultWithdrawEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeFungibleVaultWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedFungibleVaultBlueprintEventFungibleVaultDepositEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeFungibleVaultDepositEventINSTANCE.Write(writer, variant_value.Value)
		case TypedFungibleVaultBlueprintEventFungibleVaultRecallEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeFungibleVaultRecallEventINSTANCE.Write(writer, variant_value.Value)
		case TypedFungibleVaultBlueprintEventFungibleVaultPayFeeEventValue:
			writeInt32(writer, 5)
			FfiConverterTypeFungibleVaultPayFeeEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedFungibleVaultBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedFungibleVaultBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedFungibleVaultBlueprintEvent) Destroy(value TypedFungibleVaultBlueprintEvent) {
	value.Destroy()
}




type TypedLockerPackageEvent interface {
	Destroy()
}
type TypedLockerPackageEventAccountLocker struct {
	Value TypedAccountLockerBlueprintEvent
}

func (e TypedLockerPackageEventAccountLocker) Destroy() {
		FfiDestroyerTypeTypedAccountLockerBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedLockerPackageEvent struct {}

var FfiConverterTypeTypedLockerPackageEventINSTANCE = FfiConverterTypeTypedLockerPackageEvent{}

func (c FfiConverterTypeTypedLockerPackageEvent) Lift(rb RustBufferI) TypedLockerPackageEvent {
	return LiftFromRustBuffer[TypedLockerPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedLockerPackageEvent) Lower(value TypedLockerPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedLockerPackageEvent](c, value)
}
func (FfiConverterTypeTypedLockerPackageEvent) Read(reader io.Reader) TypedLockerPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedLockerPackageEventAccountLocker{
				FfiConverterTypeTypedAccountLockerBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedLockerPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedLockerPackageEvent) Write(writer io.Writer, value TypedLockerPackageEvent) {
	switch variant_value := value.(type) {
		case TypedLockerPackageEventAccountLocker:
			writeInt32(writer, 1)
			FfiConverterTypeTypedAccountLockerBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedLockerPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedLockerPackageEvent struct {}

func (_ FfiDestroyerTypeTypedLockerPackageEvent) Destroy(value TypedLockerPackageEvent) {
	value.Destroy()
}




type TypedMetadataBlueprintEvent interface {
	Destroy()
}
type TypedMetadataBlueprintEventSetMetadataEventValue struct {
	Value SetMetadataEvent
}

func (e TypedMetadataBlueprintEventSetMetadataEventValue) Destroy() {
		FfiDestroyerTypeSetMetadataEvent{}.Destroy(e.Value);
}
type TypedMetadataBlueprintEventRemoveMetadataEventValue struct {
	Value RemoveMetadataEvent
}

func (e TypedMetadataBlueprintEventRemoveMetadataEventValue) Destroy() {
		FfiDestroyerTypeRemoveMetadataEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedMetadataBlueprintEvent struct {}

var FfiConverterTypeTypedMetadataBlueprintEventINSTANCE = FfiConverterTypeTypedMetadataBlueprintEvent{}

func (c FfiConverterTypeTypedMetadataBlueprintEvent) Lift(rb RustBufferI) TypedMetadataBlueprintEvent {
	return LiftFromRustBuffer[TypedMetadataBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedMetadataBlueprintEvent) Lower(value TypedMetadataBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedMetadataBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedMetadataBlueprintEvent) Read(reader io.Reader) TypedMetadataBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedMetadataBlueprintEventSetMetadataEventValue{
				FfiConverterTypeSetMetadataEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedMetadataBlueprintEventRemoveMetadataEventValue{
				FfiConverterTypeRemoveMetadataEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedMetadataBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedMetadataBlueprintEvent) Write(writer io.Writer, value TypedMetadataBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedMetadataBlueprintEventSetMetadataEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeSetMetadataEventINSTANCE.Write(writer, variant_value.Value)
		case TypedMetadataBlueprintEventRemoveMetadataEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeRemoveMetadataEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedMetadataBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedMetadataBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedMetadataBlueprintEvent) Destroy(value TypedMetadataBlueprintEvent) {
	value.Destroy()
}




type TypedMetadataPackageEvent interface {
	Destroy()
}
type TypedMetadataPackageEventMetadata struct {
	Value TypedMetadataBlueprintEvent
}

func (e TypedMetadataPackageEventMetadata) Destroy() {
		FfiDestroyerTypeTypedMetadataBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedMetadataPackageEvent struct {}

var FfiConverterTypeTypedMetadataPackageEventINSTANCE = FfiConverterTypeTypedMetadataPackageEvent{}

func (c FfiConverterTypeTypedMetadataPackageEvent) Lift(rb RustBufferI) TypedMetadataPackageEvent {
	return LiftFromRustBuffer[TypedMetadataPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedMetadataPackageEvent) Lower(value TypedMetadataPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedMetadataPackageEvent](c, value)
}
func (FfiConverterTypeTypedMetadataPackageEvent) Read(reader io.Reader) TypedMetadataPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedMetadataPackageEventMetadata{
				FfiConverterTypeTypedMetadataBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedMetadataPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedMetadataPackageEvent) Write(writer io.Writer, value TypedMetadataPackageEvent) {
	switch variant_value := value.(type) {
		case TypedMetadataPackageEventMetadata:
			writeInt32(writer, 1)
			FfiConverterTypeTypedMetadataBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedMetadataPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedMetadataPackageEvent struct {}

func (_ FfiDestroyerTypeTypedMetadataPackageEvent) Destroy(value TypedMetadataPackageEvent) {
	value.Destroy()
}




type TypedMultiResourcePoolBlueprintEvent interface {
	Destroy()
}
type TypedMultiResourcePoolBlueprintEventMultiResourcePoolContributionEventValue struct {
	Value MultiResourcePoolContributionEvent
}

func (e TypedMultiResourcePoolBlueprintEventMultiResourcePoolContributionEventValue) Destroy() {
		FfiDestroyerTypeMultiResourcePoolContributionEvent{}.Destroy(e.Value);
}
type TypedMultiResourcePoolBlueprintEventMultiResourcePoolRedemptionEventValue struct {
	Value MultiResourcePoolRedemptionEvent
}

func (e TypedMultiResourcePoolBlueprintEventMultiResourcePoolRedemptionEventValue) Destroy() {
		FfiDestroyerTypeMultiResourcePoolRedemptionEvent{}.Destroy(e.Value);
}
type TypedMultiResourcePoolBlueprintEventMultiResourcePoolWithdrawEventValue struct {
	Value MultiResourcePoolWithdrawEvent
}

func (e TypedMultiResourcePoolBlueprintEventMultiResourcePoolWithdrawEventValue) Destroy() {
		FfiDestroyerTypeMultiResourcePoolWithdrawEvent{}.Destroy(e.Value);
}
type TypedMultiResourcePoolBlueprintEventMultiResourcePoolDepositEventValue struct {
	Value MultiResourcePoolDepositEvent
}

func (e TypedMultiResourcePoolBlueprintEventMultiResourcePoolDepositEventValue) Destroy() {
		FfiDestroyerTypeMultiResourcePoolDepositEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedMultiResourcePoolBlueprintEvent struct {}

var FfiConverterTypeTypedMultiResourcePoolBlueprintEventINSTANCE = FfiConverterTypeTypedMultiResourcePoolBlueprintEvent{}

func (c FfiConverterTypeTypedMultiResourcePoolBlueprintEvent) Lift(rb RustBufferI) TypedMultiResourcePoolBlueprintEvent {
	return LiftFromRustBuffer[TypedMultiResourcePoolBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedMultiResourcePoolBlueprintEvent) Lower(value TypedMultiResourcePoolBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedMultiResourcePoolBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedMultiResourcePoolBlueprintEvent) Read(reader io.Reader) TypedMultiResourcePoolBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedMultiResourcePoolBlueprintEventMultiResourcePoolContributionEventValue{
				FfiConverterTypeMultiResourcePoolContributionEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedMultiResourcePoolBlueprintEventMultiResourcePoolRedemptionEventValue{
				FfiConverterTypeMultiResourcePoolRedemptionEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedMultiResourcePoolBlueprintEventMultiResourcePoolWithdrawEventValue{
				FfiConverterTypeMultiResourcePoolWithdrawEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedMultiResourcePoolBlueprintEventMultiResourcePoolDepositEventValue{
				FfiConverterTypeMultiResourcePoolDepositEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedMultiResourcePoolBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedMultiResourcePoolBlueprintEvent) Write(writer io.Writer, value TypedMultiResourcePoolBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedMultiResourcePoolBlueprintEventMultiResourcePoolContributionEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeMultiResourcePoolContributionEventINSTANCE.Write(writer, variant_value.Value)
		case TypedMultiResourcePoolBlueprintEventMultiResourcePoolRedemptionEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeMultiResourcePoolRedemptionEventINSTANCE.Write(writer, variant_value.Value)
		case TypedMultiResourcePoolBlueprintEventMultiResourcePoolWithdrawEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeMultiResourcePoolWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedMultiResourcePoolBlueprintEventMultiResourcePoolDepositEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeMultiResourcePoolDepositEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedMultiResourcePoolBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedMultiResourcePoolBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedMultiResourcePoolBlueprintEvent) Destroy(value TypedMultiResourcePoolBlueprintEvent) {
	value.Destroy()
}




type TypedNativeEvent interface {
	Destroy()
}
type TypedNativeEventAccessController struct {
	Value TypedAccessControllerPackageEvent
}

func (e TypedNativeEventAccessController) Destroy() {
		FfiDestroyerTypeTypedAccessControllerPackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventAccount struct {
	Value TypedAccountPackageEvent
}

func (e TypedNativeEventAccount) Destroy() {
		FfiDestroyerTypeTypedAccountPackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventConsensusManager struct {
	Value TypedConsensusManagerPackageEvent
}

func (e TypedNativeEventConsensusManager) Destroy() {
		FfiDestroyerTypeTypedConsensusManagerPackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventPool struct {
	Value TypedPoolPackageEvent
}

func (e TypedNativeEventPool) Destroy() {
		FfiDestroyerTypeTypedPoolPackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventResource struct {
	Value TypedResourcePackageEvent
}

func (e TypedNativeEventResource) Destroy() {
		FfiDestroyerTypeTypedResourcePackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventRoleAssignment struct {
	Value TypedRoleAssignmentPackageEvent
}

func (e TypedNativeEventRoleAssignment) Destroy() {
		FfiDestroyerTypeTypedRoleAssignmentPackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventMetadata struct {
	Value TypedMetadataPackageEvent
}

func (e TypedNativeEventMetadata) Destroy() {
		FfiDestroyerTypeTypedMetadataPackageEvent{}.Destroy(e.Value);
}
type TypedNativeEventLocker struct {
	Value TypedLockerPackageEvent
}

func (e TypedNativeEventLocker) Destroy() {
		FfiDestroyerTypeTypedLockerPackageEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedNativeEvent struct {}

var FfiConverterTypeTypedNativeEventINSTANCE = FfiConverterTypeTypedNativeEvent{}

func (c FfiConverterTypeTypedNativeEvent) Lift(rb RustBufferI) TypedNativeEvent {
	return LiftFromRustBuffer[TypedNativeEvent](c, rb)
}

func (c FfiConverterTypeTypedNativeEvent) Lower(value TypedNativeEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedNativeEvent](c, value)
}
func (FfiConverterTypeTypedNativeEvent) Read(reader io.Reader) TypedNativeEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedNativeEventAccessController{
				FfiConverterTypeTypedAccessControllerPackageEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedNativeEventAccount{
				FfiConverterTypeTypedAccountPackageEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedNativeEventConsensusManager{
				FfiConverterTypeTypedConsensusManagerPackageEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedNativeEventPool{
				FfiConverterTypeTypedPoolPackageEventINSTANCE.Read(reader),
			};
		case 5:
			return TypedNativeEventResource{
				FfiConverterTypeTypedResourcePackageEventINSTANCE.Read(reader),
			};
		case 6:
			return TypedNativeEventRoleAssignment{
				FfiConverterTypeTypedRoleAssignmentPackageEventINSTANCE.Read(reader),
			};
		case 7:
			return TypedNativeEventMetadata{
				FfiConverterTypeTypedMetadataPackageEventINSTANCE.Read(reader),
			};
		case 8:
			return TypedNativeEventLocker{
				FfiConverterTypeTypedLockerPackageEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedNativeEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedNativeEvent) Write(writer io.Writer, value TypedNativeEvent) {
	switch variant_value := value.(type) {
		case TypedNativeEventAccessController:
			writeInt32(writer, 1)
			FfiConverterTypeTypedAccessControllerPackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventAccount:
			writeInt32(writer, 2)
			FfiConverterTypeTypedAccountPackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventConsensusManager:
			writeInt32(writer, 3)
			FfiConverterTypeTypedConsensusManagerPackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventPool:
			writeInt32(writer, 4)
			FfiConverterTypeTypedPoolPackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventResource:
			writeInt32(writer, 5)
			FfiConverterTypeTypedResourcePackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventRoleAssignment:
			writeInt32(writer, 6)
			FfiConverterTypeTypedRoleAssignmentPackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventMetadata:
			writeInt32(writer, 7)
			FfiConverterTypeTypedMetadataPackageEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNativeEventLocker:
			writeInt32(writer, 8)
			FfiConverterTypeTypedLockerPackageEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedNativeEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedNativeEvent struct {}

func (_ FfiDestroyerTypeTypedNativeEvent) Destroy(value TypedNativeEvent) {
	value.Destroy()
}




type TypedNonFungibleResourceManagerBlueprintEvent interface {
	Destroy()
}
type TypedNonFungibleResourceManagerBlueprintEventVaultCreationEventValue struct {
	Value VaultCreationEvent
}

func (e TypedNonFungibleResourceManagerBlueprintEventVaultCreationEventValue) Destroy() {
		FfiDestroyerTypeVaultCreationEvent{}.Destroy(e.Value);
}
type TypedNonFungibleResourceManagerBlueprintEventMintNonFungibleResourceEventValue struct {
	Value MintNonFungibleResourceEvent
}

func (e TypedNonFungibleResourceManagerBlueprintEventMintNonFungibleResourceEventValue) Destroy() {
		FfiDestroyerTypeMintNonFungibleResourceEvent{}.Destroy(e.Value);
}
type TypedNonFungibleResourceManagerBlueprintEventBurnNonFungibleResourceEventValue struct {
	Value BurnNonFungibleResourceEvent
}

func (e TypedNonFungibleResourceManagerBlueprintEventBurnNonFungibleResourceEventValue) Destroy() {
		FfiDestroyerTypeBurnNonFungibleResourceEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent struct {}

var FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEventINSTANCE = FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent{}

func (c FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent) Lift(rb RustBufferI) TypedNonFungibleResourceManagerBlueprintEvent {
	return LiftFromRustBuffer[TypedNonFungibleResourceManagerBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent) Lower(value TypedNonFungibleResourceManagerBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedNonFungibleResourceManagerBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent) Read(reader io.Reader) TypedNonFungibleResourceManagerBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedNonFungibleResourceManagerBlueprintEventVaultCreationEventValue{
				FfiConverterTypeVaultCreationEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedNonFungibleResourceManagerBlueprintEventMintNonFungibleResourceEventValue{
				FfiConverterTypeMintNonFungibleResourceEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedNonFungibleResourceManagerBlueprintEventBurnNonFungibleResourceEventValue{
				FfiConverterTypeBurnNonFungibleResourceEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent) Write(writer io.Writer, value TypedNonFungibleResourceManagerBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedNonFungibleResourceManagerBlueprintEventVaultCreationEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeVaultCreationEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNonFungibleResourceManagerBlueprintEventMintNonFungibleResourceEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeMintNonFungibleResourceEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNonFungibleResourceManagerBlueprintEventBurnNonFungibleResourceEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeBurnNonFungibleResourceEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedNonFungibleResourceManagerBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedNonFungibleResourceManagerBlueprintEvent) Destroy(value TypedNonFungibleResourceManagerBlueprintEvent) {
	value.Destroy()
}




type TypedNonFungibleVaultBlueprintEvent interface {
	Destroy()
}
type TypedNonFungibleVaultBlueprintEventNonFungibleVaultWithdrawEventValue struct {
	Value NonFungibleVaultWithdrawEvent
}

func (e TypedNonFungibleVaultBlueprintEventNonFungibleVaultWithdrawEventValue) Destroy() {
		FfiDestroyerTypeNonFungibleVaultWithdrawEvent{}.Destroy(e.Value);
}
type TypedNonFungibleVaultBlueprintEventNonFungibleVaultDepositEventValue struct {
	Value NonFungibleVaultDepositEvent
}

func (e TypedNonFungibleVaultBlueprintEventNonFungibleVaultDepositEventValue) Destroy() {
		FfiDestroyerTypeNonFungibleVaultDepositEvent{}.Destroy(e.Value);
}
type TypedNonFungibleVaultBlueprintEventNonFungibleVaultRecallEventValue struct {
	Value NonFungibleVaultRecallEvent
}

func (e TypedNonFungibleVaultBlueprintEventNonFungibleVaultRecallEventValue) Destroy() {
		FfiDestroyerTypeNonFungibleVaultRecallEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedNonFungibleVaultBlueprintEvent struct {}

var FfiConverterTypeTypedNonFungibleVaultBlueprintEventINSTANCE = FfiConverterTypeTypedNonFungibleVaultBlueprintEvent{}

func (c FfiConverterTypeTypedNonFungibleVaultBlueprintEvent) Lift(rb RustBufferI) TypedNonFungibleVaultBlueprintEvent {
	return LiftFromRustBuffer[TypedNonFungibleVaultBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedNonFungibleVaultBlueprintEvent) Lower(value TypedNonFungibleVaultBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedNonFungibleVaultBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedNonFungibleVaultBlueprintEvent) Read(reader io.Reader) TypedNonFungibleVaultBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedNonFungibleVaultBlueprintEventNonFungibleVaultWithdrawEventValue{
				FfiConverterTypeNonFungibleVaultWithdrawEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedNonFungibleVaultBlueprintEventNonFungibleVaultDepositEventValue{
				FfiConverterTypeNonFungibleVaultDepositEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedNonFungibleVaultBlueprintEventNonFungibleVaultRecallEventValue{
				FfiConverterTypeNonFungibleVaultRecallEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedNonFungibleVaultBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedNonFungibleVaultBlueprintEvent) Write(writer io.Writer, value TypedNonFungibleVaultBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedNonFungibleVaultBlueprintEventNonFungibleVaultWithdrawEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeNonFungibleVaultWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNonFungibleVaultBlueprintEventNonFungibleVaultDepositEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeNonFungibleVaultDepositEventINSTANCE.Write(writer, variant_value.Value)
		case TypedNonFungibleVaultBlueprintEventNonFungibleVaultRecallEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeNonFungibleVaultRecallEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedNonFungibleVaultBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedNonFungibleVaultBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedNonFungibleVaultBlueprintEvent) Destroy(value TypedNonFungibleVaultBlueprintEvent) {
	value.Destroy()
}




type TypedOneResourcePoolBlueprintEvent interface {
	Destroy()
}
type TypedOneResourcePoolBlueprintEventOneResourcePoolContributionEventValue struct {
	Value OneResourcePoolContributionEvent
}

func (e TypedOneResourcePoolBlueprintEventOneResourcePoolContributionEventValue) Destroy() {
		FfiDestroyerTypeOneResourcePoolContributionEvent{}.Destroy(e.Value);
}
type TypedOneResourcePoolBlueprintEventOneResourcePoolRedemptionEventValue struct {
	Value OneResourcePoolRedemptionEvent
}

func (e TypedOneResourcePoolBlueprintEventOneResourcePoolRedemptionEventValue) Destroy() {
		FfiDestroyerTypeOneResourcePoolRedemptionEvent{}.Destroy(e.Value);
}
type TypedOneResourcePoolBlueprintEventOneResourcePoolWithdrawEventValue struct {
	Value OneResourcePoolWithdrawEvent
}

func (e TypedOneResourcePoolBlueprintEventOneResourcePoolWithdrawEventValue) Destroy() {
		FfiDestroyerTypeOneResourcePoolWithdrawEvent{}.Destroy(e.Value);
}
type TypedOneResourcePoolBlueprintEventOneResourcePoolDepositEventValue struct {
	Value OneResourcePoolDepositEvent
}

func (e TypedOneResourcePoolBlueprintEventOneResourcePoolDepositEventValue) Destroy() {
		FfiDestroyerTypeOneResourcePoolDepositEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedOneResourcePoolBlueprintEvent struct {}

var FfiConverterTypeTypedOneResourcePoolBlueprintEventINSTANCE = FfiConverterTypeTypedOneResourcePoolBlueprintEvent{}

func (c FfiConverterTypeTypedOneResourcePoolBlueprintEvent) Lift(rb RustBufferI) TypedOneResourcePoolBlueprintEvent {
	return LiftFromRustBuffer[TypedOneResourcePoolBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedOneResourcePoolBlueprintEvent) Lower(value TypedOneResourcePoolBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedOneResourcePoolBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedOneResourcePoolBlueprintEvent) Read(reader io.Reader) TypedOneResourcePoolBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedOneResourcePoolBlueprintEventOneResourcePoolContributionEventValue{
				FfiConverterTypeOneResourcePoolContributionEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedOneResourcePoolBlueprintEventOneResourcePoolRedemptionEventValue{
				FfiConverterTypeOneResourcePoolRedemptionEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedOneResourcePoolBlueprintEventOneResourcePoolWithdrawEventValue{
				FfiConverterTypeOneResourcePoolWithdrawEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedOneResourcePoolBlueprintEventOneResourcePoolDepositEventValue{
				FfiConverterTypeOneResourcePoolDepositEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedOneResourcePoolBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedOneResourcePoolBlueprintEvent) Write(writer io.Writer, value TypedOneResourcePoolBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedOneResourcePoolBlueprintEventOneResourcePoolContributionEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeOneResourcePoolContributionEventINSTANCE.Write(writer, variant_value.Value)
		case TypedOneResourcePoolBlueprintEventOneResourcePoolRedemptionEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeOneResourcePoolRedemptionEventINSTANCE.Write(writer, variant_value.Value)
		case TypedOneResourcePoolBlueprintEventOneResourcePoolWithdrawEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeOneResourcePoolWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedOneResourcePoolBlueprintEventOneResourcePoolDepositEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeOneResourcePoolDepositEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedOneResourcePoolBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedOneResourcePoolBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedOneResourcePoolBlueprintEvent) Destroy(value TypedOneResourcePoolBlueprintEvent) {
	value.Destroy()
}




type TypedPoolPackageEvent interface {
	Destroy()
}
type TypedPoolPackageEventOneResourcePool struct {
	Value TypedOneResourcePoolBlueprintEvent
}

func (e TypedPoolPackageEventOneResourcePool) Destroy() {
		FfiDestroyerTypeTypedOneResourcePoolBlueprintEvent{}.Destroy(e.Value);
}
type TypedPoolPackageEventTwoResourcePool struct {
	Value TypedTwoResourcePoolBlueprintEvent
}

func (e TypedPoolPackageEventTwoResourcePool) Destroy() {
		FfiDestroyerTypeTypedTwoResourcePoolBlueprintEvent{}.Destroy(e.Value);
}
type TypedPoolPackageEventMultiResourcePool struct {
	Value TypedMultiResourcePoolBlueprintEvent
}

func (e TypedPoolPackageEventMultiResourcePool) Destroy() {
		FfiDestroyerTypeTypedMultiResourcePoolBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedPoolPackageEvent struct {}

var FfiConverterTypeTypedPoolPackageEventINSTANCE = FfiConverterTypeTypedPoolPackageEvent{}

func (c FfiConverterTypeTypedPoolPackageEvent) Lift(rb RustBufferI) TypedPoolPackageEvent {
	return LiftFromRustBuffer[TypedPoolPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedPoolPackageEvent) Lower(value TypedPoolPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedPoolPackageEvent](c, value)
}
func (FfiConverterTypeTypedPoolPackageEvent) Read(reader io.Reader) TypedPoolPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedPoolPackageEventOneResourcePool{
				FfiConverterTypeTypedOneResourcePoolBlueprintEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedPoolPackageEventTwoResourcePool{
				FfiConverterTypeTypedTwoResourcePoolBlueprintEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedPoolPackageEventMultiResourcePool{
				FfiConverterTypeTypedMultiResourcePoolBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedPoolPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedPoolPackageEvent) Write(writer io.Writer, value TypedPoolPackageEvent) {
	switch variant_value := value.(type) {
		case TypedPoolPackageEventOneResourcePool:
			writeInt32(writer, 1)
			FfiConverterTypeTypedOneResourcePoolBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		case TypedPoolPackageEventTwoResourcePool:
			writeInt32(writer, 2)
			FfiConverterTypeTypedTwoResourcePoolBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		case TypedPoolPackageEventMultiResourcePool:
			writeInt32(writer, 3)
			FfiConverterTypeTypedMultiResourcePoolBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedPoolPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedPoolPackageEvent struct {}

func (_ FfiDestroyerTypeTypedPoolPackageEvent) Destroy(value TypedPoolPackageEvent) {
	value.Destroy()
}




type TypedResourcePackageEvent interface {
	Destroy()
}
type TypedResourcePackageEventFungibleVault struct {
	Value TypedFungibleVaultBlueprintEvent
}

func (e TypedResourcePackageEventFungibleVault) Destroy() {
		FfiDestroyerTypeTypedFungibleVaultBlueprintEvent{}.Destroy(e.Value);
}
type TypedResourcePackageEventNonFungibleVault struct {
	Value TypedNonFungibleVaultBlueprintEvent
}

func (e TypedResourcePackageEventNonFungibleVault) Destroy() {
		FfiDestroyerTypeTypedNonFungibleVaultBlueprintEvent{}.Destroy(e.Value);
}
type TypedResourcePackageEventFungibleResourceManager struct {
	Value TypedFungibleResourceManagerBlueprintEvent
}

func (e TypedResourcePackageEventFungibleResourceManager) Destroy() {
		FfiDestroyerTypeTypedFungibleResourceManagerBlueprintEvent{}.Destroy(e.Value);
}
type TypedResourcePackageEventNonFungibleResourceManager struct {
	Value TypedNonFungibleResourceManagerBlueprintEvent
}

func (e TypedResourcePackageEventNonFungibleResourceManager) Destroy() {
		FfiDestroyerTypeTypedNonFungibleResourceManagerBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedResourcePackageEvent struct {}

var FfiConverterTypeTypedResourcePackageEventINSTANCE = FfiConverterTypeTypedResourcePackageEvent{}

func (c FfiConverterTypeTypedResourcePackageEvent) Lift(rb RustBufferI) TypedResourcePackageEvent {
	return LiftFromRustBuffer[TypedResourcePackageEvent](c, rb)
}

func (c FfiConverterTypeTypedResourcePackageEvent) Lower(value TypedResourcePackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedResourcePackageEvent](c, value)
}
func (FfiConverterTypeTypedResourcePackageEvent) Read(reader io.Reader) TypedResourcePackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedResourcePackageEventFungibleVault{
				FfiConverterTypeTypedFungibleVaultBlueprintEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedResourcePackageEventNonFungibleVault{
				FfiConverterTypeTypedNonFungibleVaultBlueprintEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedResourcePackageEventFungibleResourceManager{
				FfiConverterTypeTypedFungibleResourceManagerBlueprintEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedResourcePackageEventNonFungibleResourceManager{
				FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedResourcePackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedResourcePackageEvent) Write(writer io.Writer, value TypedResourcePackageEvent) {
	switch variant_value := value.(type) {
		case TypedResourcePackageEventFungibleVault:
			writeInt32(writer, 1)
			FfiConverterTypeTypedFungibleVaultBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		case TypedResourcePackageEventNonFungibleVault:
			writeInt32(writer, 2)
			FfiConverterTypeTypedNonFungibleVaultBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		case TypedResourcePackageEventFungibleResourceManager:
			writeInt32(writer, 3)
			FfiConverterTypeTypedFungibleResourceManagerBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		case TypedResourcePackageEventNonFungibleResourceManager:
			writeInt32(writer, 4)
			FfiConverterTypeTypedNonFungibleResourceManagerBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedResourcePackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedResourcePackageEvent struct {}

func (_ FfiDestroyerTypeTypedResourcePackageEvent) Destroy(value TypedResourcePackageEvent) {
	value.Destroy()
}




type TypedRoleAssignmentBlueprintEvent interface {
	Destroy()
}
type TypedRoleAssignmentBlueprintEventSetRoleEventValue struct {
	Value SetRoleEvent
}

func (e TypedRoleAssignmentBlueprintEventSetRoleEventValue) Destroy() {
		FfiDestroyerTypeSetRoleEvent{}.Destroy(e.Value);
}
type TypedRoleAssignmentBlueprintEventSetOwnerRoleEventValue struct {
	Value SetOwnerRoleEvent
}

func (e TypedRoleAssignmentBlueprintEventSetOwnerRoleEventValue) Destroy() {
		FfiDestroyerTypeSetOwnerRoleEvent{}.Destroy(e.Value);
}
type TypedRoleAssignmentBlueprintEventLockOwnerRoleEventValue struct {
	Value LockOwnerRoleEvent
}

func (e TypedRoleAssignmentBlueprintEventLockOwnerRoleEventValue) Destroy() {
		FfiDestroyerTypeLockOwnerRoleEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedRoleAssignmentBlueprintEvent struct {}

var FfiConverterTypeTypedRoleAssignmentBlueprintEventINSTANCE = FfiConverterTypeTypedRoleAssignmentBlueprintEvent{}

func (c FfiConverterTypeTypedRoleAssignmentBlueprintEvent) Lift(rb RustBufferI) TypedRoleAssignmentBlueprintEvent {
	return LiftFromRustBuffer[TypedRoleAssignmentBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedRoleAssignmentBlueprintEvent) Lower(value TypedRoleAssignmentBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedRoleAssignmentBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedRoleAssignmentBlueprintEvent) Read(reader io.Reader) TypedRoleAssignmentBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedRoleAssignmentBlueprintEventSetRoleEventValue{
				FfiConverterTypeSetRoleEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedRoleAssignmentBlueprintEventSetOwnerRoleEventValue{
				FfiConverterTypeSetOwnerRoleEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedRoleAssignmentBlueprintEventLockOwnerRoleEventValue{
				FfiConverterTypeLockOwnerRoleEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedRoleAssignmentBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedRoleAssignmentBlueprintEvent) Write(writer io.Writer, value TypedRoleAssignmentBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedRoleAssignmentBlueprintEventSetRoleEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeSetRoleEventINSTANCE.Write(writer, variant_value.Value)
		case TypedRoleAssignmentBlueprintEventSetOwnerRoleEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeSetOwnerRoleEventINSTANCE.Write(writer, variant_value.Value)
		case TypedRoleAssignmentBlueprintEventLockOwnerRoleEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeLockOwnerRoleEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedRoleAssignmentBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedRoleAssignmentBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedRoleAssignmentBlueprintEvent) Destroy(value TypedRoleAssignmentBlueprintEvent) {
	value.Destroy()
}




type TypedRoleAssignmentPackageEvent interface {
	Destroy()
}
type TypedRoleAssignmentPackageEventRoleAssignment struct {
	Value TypedRoleAssignmentBlueprintEvent
}

func (e TypedRoleAssignmentPackageEventRoleAssignment) Destroy() {
		FfiDestroyerTypeTypedRoleAssignmentBlueprintEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedRoleAssignmentPackageEvent struct {}

var FfiConverterTypeTypedRoleAssignmentPackageEventINSTANCE = FfiConverterTypeTypedRoleAssignmentPackageEvent{}

func (c FfiConverterTypeTypedRoleAssignmentPackageEvent) Lift(rb RustBufferI) TypedRoleAssignmentPackageEvent {
	return LiftFromRustBuffer[TypedRoleAssignmentPackageEvent](c, rb)
}

func (c FfiConverterTypeTypedRoleAssignmentPackageEvent) Lower(value TypedRoleAssignmentPackageEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedRoleAssignmentPackageEvent](c, value)
}
func (FfiConverterTypeTypedRoleAssignmentPackageEvent) Read(reader io.Reader) TypedRoleAssignmentPackageEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedRoleAssignmentPackageEventRoleAssignment{
				FfiConverterTypeTypedRoleAssignmentBlueprintEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedRoleAssignmentPackageEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedRoleAssignmentPackageEvent) Write(writer io.Writer, value TypedRoleAssignmentPackageEvent) {
	switch variant_value := value.(type) {
		case TypedRoleAssignmentPackageEventRoleAssignment:
			writeInt32(writer, 1)
			FfiConverterTypeTypedRoleAssignmentBlueprintEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedRoleAssignmentPackageEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedRoleAssignmentPackageEvent struct {}

func (_ FfiDestroyerTypeTypedRoleAssignmentPackageEvent) Destroy(value TypedRoleAssignmentPackageEvent) {
	value.Destroy()
}




type TypedTwoResourcePoolBlueprintEvent interface {
	Destroy()
}
type TypedTwoResourcePoolBlueprintEventTwoResourcePoolContributionEventValue struct {
	Value TwoResourcePoolContributionEvent
}

func (e TypedTwoResourcePoolBlueprintEventTwoResourcePoolContributionEventValue) Destroy() {
		FfiDestroyerTypeTwoResourcePoolContributionEvent{}.Destroy(e.Value);
}
type TypedTwoResourcePoolBlueprintEventTwoResourcePoolRedemptionEventValue struct {
	Value TwoResourcePoolRedemptionEvent
}

func (e TypedTwoResourcePoolBlueprintEventTwoResourcePoolRedemptionEventValue) Destroy() {
		FfiDestroyerTypeTwoResourcePoolRedemptionEvent{}.Destroy(e.Value);
}
type TypedTwoResourcePoolBlueprintEventTwoResourcePoolWithdrawEventValue struct {
	Value TwoResourcePoolWithdrawEvent
}

func (e TypedTwoResourcePoolBlueprintEventTwoResourcePoolWithdrawEventValue) Destroy() {
		FfiDestroyerTypeTwoResourcePoolWithdrawEvent{}.Destroy(e.Value);
}
type TypedTwoResourcePoolBlueprintEventTwoResourcePoolDepositEventValue struct {
	Value TwoResourcePoolDepositEvent
}

func (e TypedTwoResourcePoolBlueprintEventTwoResourcePoolDepositEventValue) Destroy() {
		FfiDestroyerTypeTwoResourcePoolDepositEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedTwoResourcePoolBlueprintEvent struct {}

var FfiConverterTypeTypedTwoResourcePoolBlueprintEventINSTANCE = FfiConverterTypeTypedTwoResourcePoolBlueprintEvent{}

func (c FfiConverterTypeTypedTwoResourcePoolBlueprintEvent) Lift(rb RustBufferI) TypedTwoResourcePoolBlueprintEvent {
	return LiftFromRustBuffer[TypedTwoResourcePoolBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedTwoResourcePoolBlueprintEvent) Lower(value TypedTwoResourcePoolBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedTwoResourcePoolBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedTwoResourcePoolBlueprintEvent) Read(reader io.Reader) TypedTwoResourcePoolBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedTwoResourcePoolBlueprintEventTwoResourcePoolContributionEventValue{
				FfiConverterTypeTwoResourcePoolContributionEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedTwoResourcePoolBlueprintEventTwoResourcePoolRedemptionEventValue{
				FfiConverterTypeTwoResourcePoolRedemptionEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedTwoResourcePoolBlueprintEventTwoResourcePoolWithdrawEventValue{
				FfiConverterTypeTwoResourcePoolWithdrawEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedTwoResourcePoolBlueprintEventTwoResourcePoolDepositEventValue{
				FfiConverterTypeTwoResourcePoolDepositEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedTwoResourcePoolBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedTwoResourcePoolBlueprintEvent) Write(writer io.Writer, value TypedTwoResourcePoolBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedTwoResourcePoolBlueprintEventTwoResourcePoolContributionEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeTwoResourcePoolContributionEventINSTANCE.Write(writer, variant_value.Value)
		case TypedTwoResourcePoolBlueprintEventTwoResourcePoolRedemptionEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeTwoResourcePoolRedemptionEventINSTANCE.Write(writer, variant_value.Value)
		case TypedTwoResourcePoolBlueprintEventTwoResourcePoolWithdrawEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeTwoResourcePoolWithdrawEventINSTANCE.Write(writer, variant_value.Value)
		case TypedTwoResourcePoolBlueprintEventTwoResourcePoolDepositEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeTwoResourcePoolDepositEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedTwoResourcePoolBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedTwoResourcePoolBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedTwoResourcePoolBlueprintEvent) Destroy(value TypedTwoResourcePoolBlueprintEvent) {
	value.Destroy()
}




type TypedValidatorBlueprintEvent interface {
	Destroy()
}
type TypedValidatorBlueprintEventRegisterValidatorEventValue struct {
	Value RegisterValidatorEvent
}

func (e TypedValidatorBlueprintEventRegisterValidatorEventValue) Destroy() {
		FfiDestroyerTypeRegisterValidatorEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventUnregisterValidatorEventValue struct {
	Value UnregisterValidatorEvent
}

func (e TypedValidatorBlueprintEventUnregisterValidatorEventValue) Destroy() {
		FfiDestroyerTypeUnregisterValidatorEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventStakeEventValue struct {
	Value StakeEvent
}

func (e TypedValidatorBlueprintEventStakeEventValue) Destroy() {
		FfiDestroyerTypeStakeEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventUnstakeEventValue struct {
	Value UnstakeEvent
}

func (e TypedValidatorBlueprintEventUnstakeEventValue) Destroy() {
		FfiDestroyerTypeUnstakeEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventClaimXrdEventValue struct {
	Value ClaimXrdEvent
}

func (e TypedValidatorBlueprintEventClaimXrdEventValue) Destroy() {
		FfiDestroyerTypeClaimXrdEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventUpdateAcceptingStakeDelegationStateEventValue struct {
	Value UpdateAcceptingStakeDelegationStateEvent
}

func (e TypedValidatorBlueprintEventUpdateAcceptingStakeDelegationStateEventValue) Destroy() {
		FfiDestroyerTypeUpdateAcceptingStakeDelegationStateEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventProtocolUpdateReadinessSignalEventValue struct {
	Value ProtocolUpdateReadinessSignalEvent
}

func (e TypedValidatorBlueprintEventProtocolUpdateReadinessSignalEventValue) Destroy() {
		FfiDestroyerTypeProtocolUpdateReadinessSignalEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventValidatorEmissionAppliedEventValue struct {
	Value ValidatorEmissionAppliedEvent
}

func (e TypedValidatorBlueprintEventValidatorEmissionAppliedEventValue) Destroy() {
		FfiDestroyerTypeValidatorEmissionAppliedEvent{}.Destroy(e.Value);
}
type TypedValidatorBlueprintEventValidatorRewardAppliedEventValue struct {
	Value ValidatorRewardAppliedEvent
}

func (e TypedValidatorBlueprintEventValidatorRewardAppliedEventValue) Destroy() {
		FfiDestroyerTypeValidatorRewardAppliedEvent{}.Destroy(e.Value);
}

type FfiConverterTypeTypedValidatorBlueprintEvent struct {}

var FfiConverterTypeTypedValidatorBlueprintEventINSTANCE = FfiConverterTypeTypedValidatorBlueprintEvent{}

func (c FfiConverterTypeTypedValidatorBlueprintEvent) Lift(rb RustBufferI) TypedValidatorBlueprintEvent {
	return LiftFromRustBuffer[TypedValidatorBlueprintEvent](c, rb)
}

func (c FfiConverterTypeTypedValidatorBlueprintEvent) Lower(value TypedValidatorBlueprintEvent) RustBuffer {
	return LowerIntoRustBuffer[TypedValidatorBlueprintEvent](c, value)
}
func (FfiConverterTypeTypedValidatorBlueprintEvent) Read(reader io.Reader) TypedValidatorBlueprintEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TypedValidatorBlueprintEventRegisterValidatorEventValue{
				FfiConverterTypeRegisterValidatorEventINSTANCE.Read(reader),
			};
		case 2:
			return TypedValidatorBlueprintEventUnregisterValidatorEventValue{
				FfiConverterTypeUnregisterValidatorEventINSTANCE.Read(reader),
			};
		case 3:
			return TypedValidatorBlueprintEventStakeEventValue{
				FfiConverterTypeStakeEventINSTANCE.Read(reader),
			};
		case 4:
			return TypedValidatorBlueprintEventUnstakeEventValue{
				FfiConverterTypeUnstakeEventINSTANCE.Read(reader),
			};
		case 5:
			return TypedValidatorBlueprintEventClaimXrdEventValue{
				FfiConverterTypeClaimXrdEventINSTANCE.Read(reader),
			};
		case 6:
			return TypedValidatorBlueprintEventUpdateAcceptingStakeDelegationStateEventValue{
				FfiConverterTypeUpdateAcceptingStakeDelegationStateEventINSTANCE.Read(reader),
			};
		case 7:
			return TypedValidatorBlueprintEventProtocolUpdateReadinessSignalEventValue{
				FfiConverterTypeProtocolUpdateReadinessSignalEventINSTANCE.Read(reader),
			};
		case 8:
			return TypedValidatorBlueprintEventValidatorEmissionAppliedEventValue{
				FfiConverterTypeValidatorEmissionAppliedEventINSTANCE.Read(reader),
			};
		case 9:
			return TypedValidatorBlueprintEventValidatorRewardAppliedEventValue{
				FfiConverterTypeValidatorRewardAppliedEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTypedValidatorBlueprintEvent.Read()", id));
	}
}

func (FfiConverterTypeTypedValidatorBlueprintEvent) Write(writer io.Writer, value TypedValidatorBlueprintEvent) {
	switch variant_value := value.(type) {
		case TypedValidatorBlueprintEventRegisterValidatorEventValue:
			writeInt32(writer, 1)
			FfiConverterTypeRegisterValidatorEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventUnregisterValidatorEventValue:
			writeInt32(writer, 2)
			FfiConverterTypeUnregisterValidatorEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventStakeEventValue:
			writeInt32(writer, 3)
			FfiConverterTypeStakeEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventUnstakeEventValue:
			writeInt32(writer, 4)
			FfiConverterTypeUnstakeEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventClaimXrdEventValue:
			writeInt32(writer, 5)
			FfiConverterTypeClaimXrdEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventUpdateAcceptingStakeDelegationStateEventValue:
			writeInt32(writer, 6)
			FfiConverterTypeUpdateAcceptingStakeDelegationStateEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventProtocolUpdateReadinessSignalEventValue:
			writeInt32(writer, 7)
			FfiConverterTypeProtocolUpdateReadinessSignalEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventValidatorEmissionAppliedEventValue:
			writeInt32(writer, 8)
			FfiConverterTypeValidatorEmissionAppliedEventINSTANCE.Write(writer, variant_value.Value)
		case TypedValidatorBlueprintEventValidatorRewardAppliedEventValue:
			writeInt32(writer, 9)
			FfiConverterTypeValidatorRewardAppliedEventINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTypedValidatorBlueprintEvent.Write", value))
	}
}

type FfiDestroyerTypeTypedValidatorBlueprintEvent struct {}

func (_ FfiDestroyerTypeTypedValidatorBlueprintEvent) Destroy(value TypedValidatorBlueprintEvent) {
	value.Destroy()
}




type WithdrawResourceEvent interface {
	Destroy()
}
type WithdrawResourceEventAmount struct {
	Value *Decimal
}

func (e WithdrawResourceEventAmount) Destroy() {
		FfiDestroyerDecimal{}.Destroy(e.Value);
}
type WithdrawResourceEventIds struct {
	Value []NonFungibleLocalId
}

func (e WithdrawResourceEventIds) Destroy() {
		FfiDestroyerSequenceTypeNonFungibleLocalId{}.Destroy(e.Value);
}

type FfiConverterTypeWithdrawResourceEvent struct {}

var FfiConverterTypeWithdrawResourceEventINSTANCE = FfiConverterTypeWithdrawResourceEvent{}

func (c FfiConverterTypeWithdrawResourceEvent) Lift(rb RustBufferI) WithdrawResourceEvent {
	return LiftFromRustBuffer[WithdrawResourceEvent](c, rb)
}

func (c FfiConverterTypeWithdrawResourceEvent) Lower(value WithdrawResourceEvent) RustBuffer {
	return LowerIntoRustBuffer[WithdrawResourceEvent](c, value)
}
func (FfiConverterTypeWithdrawResourceEvent) Read(reader io.Reader) WithdrawResourceEvent {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return WithdrawResourceEventAmount{
				FfiConverterDecimalINSTANCE.Read(reader),
			};
		case 2:
			return WithdrawResourceEventIds{
				FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeWithdrawResourceEvent.Read()", id));
	}
}

func (FfiConverterTypeWithdrawResourceEvent) Write(writer io.Writer, value WithdrawResourceEvent) {
	switch variant_value := value.(type) {
		case WithdrawResourceEventAmount:
			writeInt32(writer, 1)
			FfiConverterDecimalINSTANCE.Write(writer, variant_value.Value)
		case WithdrawResourceEventIds:
			writeInt32(writer, 2)
			FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE.Write(writer, variant_value.Value)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeWithdrawResourceEvent.Write", value))
	}
}

type FfiDestroyerTypeWithdrawResourceEvent struct {}

func (_ FfiDestroyerTypeWithdrawResourceEvent) Destroy(value WithdrawResourceEvent) {
	value.Destroy()
}




type WithdrawStrategy interface {
	Destroy()
}
type WithdrawStrategyExact struct {
}

func (e WithdrawStrategyExact) Destroy() {
}
type WithdrawStrategyRounded struct {
	RoundingMode RoundingMode
}

func (e WithdrawStrategyRounded) Destroy() {
		FfiDestroyerTypeRoundingMode{}.Destroy(e.RoundingMode);
}

type FfiConverterTypeWithdrawStrategy struct {}

var FfiConverterTypeWithdrawStrategyINSTANCE = FfiConverterTypeWithdrawStrategy{}

func (c FfiConverterTypeWithdrawStrategy) Lift(rb RustBufferI) WithdrawStrategy {
	return LiftFromRustBuffer[WithdrawStrategy](c, rb)
}

func (c FfiConverterTypeWithdrawStrategy) Lower(value WithdrawStrategy) RustBuffer {
	return LowerIntoRustBuffer[WithdrawStrategy](c, value)
}
func (FfiConverterTypeWithdrawStrategy) Read(reader io.Reader) WithdrawStrategy {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return WithdrawStrategyExact{
			};
		case 2:
			return WithdrawStrategyRounded{
				FfiConverterTypeRoundingModeINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeWithdrawStrategy.Read()", id));
	}
}

func (FfiConverterTypeWithdrawStrategy) Write(writer io.Writer, value WithdrawStrategy) {
	switch variant_value := value.(type) {
		case WithdrawStrategyExact:
			writeInt32(writer, 1)
		case WithdrawStrategyRounded:
			writeInt32(writer, 2)
			FfiConverterTypeRoundingModeINSTANCE.Write(writer, variant_value.RoundingMode)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeWithdrawStrategy.Write", value))
	}
}

type FfiDestroyerTypeWithdrawStrategy struct {}

func (_ FfiDestroyerTypeWithdrawStrategy) Destroy(value WithdrawStrategy) {
	value.Destroy()
}




type uniffiCallbackResult C.int32_t

const (
	uniffiIdxCallbackFree               uniffiCallbackResult = 0
	uniffiCallbackResultSuccess         uniffiCallbackResult = 0
	uniffiCallbackResultError           uniffiCallbackResult = 1
	uniffiCallbackUnexpectedResultError uniffiCallbackResult = 2
	uniffiCallbackCancelled             uniffiCallbackResult = 3
)

type concurrentHandleMap[T any] struct {
	leftMap       map[uint64]*T
	rightMap      map[*T]uint64
	currentHandle uint64
	lock          sync.RWMutex
}

func newConcurrentHandleMap[T any]() *concurrentHandleMap[T] {
	return &concurrentHandleMap[T]{
		leftMap:  map[uint64]*T{},
		rightMap: map[*T]uint64{},
	}
}

func (cm *concurrentHandleMap[T]) insert(obj *T) uint64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if existingHandle, ok := cm.rightMap[obj]; ok {
		return existingHandle
	}
	cm.currentHandle = cm.currentHandle + 1
	cm.leftMap[cm.currentHandle] = obj
	cm.rightMap[obj] = cm.currentHandle
	return cm.currentHandle
}

func (cm *concurrentHandleMap[T]) remove(handle uint64) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if val, ok := cm.leftMap[handle]; ok {
		delete(cm.leftMap, handle)
		delete(cm.rightMap, val)
	}
	return false
}

func (cm *concurrentHandleMap[T]) tryGet(handle uint64) (*T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.leftMap[handle]
	return val, ok
}

type FfiConverterCallbackInterface[CallbackInterface any] struct {
	handleMap *concurrentHandleMap[CallbackInterface]
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) drop(handle uint64) RustBuffer {
	c.handleMap.remove(handle)
	return RustBuffer{}
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lift(handle uint64) CallbackInterface {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return *val
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Read(reader io.Reader) CallbackInterface {
	return c.Lift(readUint64(reader))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lower(value CallbackInterface) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(&value))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Write(writer io.Writer, value CallbackInterface) {
	writeUint64(writer, uint64(c.Lower(value)))
}
type Signer interface {
	
	Sign(hash *Hash) []byte
	
	SignToSignature(hash *Hash) Signature
	
	SignToSignatureWithPublicKey(hash *Hash) SignatureWithPublicKey
	
	PublicKey() PublicKey
	
}

// foreignCallbackCallbackInterfaceSigner cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceSigner struct {}

//export radix_engine_toolkit_uniffi_cgo_Signer
func radix_engine_toolkit_uniffi_cgo_Signer(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceSignerINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceSignerINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceSigner{}.InvokeSign(cb, args, outBuf);
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceSigner{}.InvokeSignToSignature(cb, args, outBuf);
		return C.int32_t(result)
	case 3:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceSigner{}.InvokeSignToSignatureWithPublicKey(cb, args, outBuf);
		return C.int32_t(result)
	case 4:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceSigner{}.InvokePublicKey(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceSigner) InvokeSign (callback Signer, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result :=callback.Sign(FfiConverterHashINSTANCE.Read(reader));

        
	*outBuf = LowerIntoRustBuffer[[]byte](FfiConverterBytesINSTANCE, result)
	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceSigner) InvokeSignToSignature (callback Signer, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result :=callback.SignToSignature(FfiConverterHashINSTANCE.Read(reader));

        
	*outBuf = LowerIntoRustBuffer[Signature](FfiConverterTypeSignatureINSTANCE, result)
	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceSigner) InvokeSignToSignatureWithPublicKey (callback Signer, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result :=callback.SignToSignatureWithPublicKey(FfiConverterHashINSTANCE.Read(reader));

        
	*outBuf = LowerIntoRustBuffer[SignatureWithPublicKey](FfiConverterTypeSignatureWithPublicKeyINSTANCE, result)
	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceSigner) InvokePublicKey (callback Signer, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	result :=callback.PublicKey();

        
	*outBuf = LowerIntoRustBuffer[PublicKey](FfiConverterTypePublicKeyINSTANCE, result)
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceSigner struct {
	FfiConverterCallbackInterface[Signer]
}

var FfiConverterCallbackInterfaceSignerINSTANCE = &FfiConverterCallbackInterfaceSigner {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[Signer]{
		handleMap: newConcurrentHandleMap[Signer](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceSigner) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_radix_engine_toolkit_uniffi_fn_init_callback_signer(C.ForeignCallback(C.radix_engine_toolkit_uniffi_cgo_Signer), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceSigner struct {}

func (FfiDestroyerCallbackInterfaceSigner) Destroy(value Signer) {
}



type FfiConverterOptionalUint32 struct{}

var FfiConverterOptionalUint32INSTANCE = FfiConverterOptionalUint32{}

func (c FfiConverterOptionalUint32) Lift(rb RustBufferI) *uint32 {
	return LiftFromRustBuffer[*uint32](c, rb)
}

func (_ FfiConverterOptionalUint32) Read(reader io.Reader) *uint32 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint32INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint32) Lower(value *uint32) RustBuffer {
	return LowerIntoRustBuffer[*uint32](c, value)
}

func (_ FfiConverterOptionalUint32) Write(writer io.Writer, value *uint32) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint32INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint32 struct {}

func (_ FfiDestroyerOptionalUint32) Destroy(value *uint32) {
	if value != nil {
		FfiDestroyerUint32{}.Destroy(*value)
	}
}



type FfiConverterOptionalAccessRule struct{}

var FfiConverterOptionalAccessRuleINSTANCE = FfiConverterOptionalAccessRule{}

func (c FfiConverterOptionalAccessRule) Lift(rb RustBufferI) **AccessRule {
	return LiftFromRustBuffer[**AccessRule](c, rb)
}

func (_ FfiConverterOptionalAccessRule) Read(reader io.Reader) **AccessRule {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterAccessRuleINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalAccessRule) Lower(value **AccessRule) RustBuffer {
	return LowerIntoRustBuffer[**AccessRule](c, value)
}

func (_ FfiConverterOptionalAccessRule) Write(writer io.Writer, value **AccessRule) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterAccessRuleINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalAccessRule struct {}

func (_ FfiDestroyerOptionalAccessRule) Destroy(value **AccessRule) {
	if value != nil {
		FfiDestroyerAccessRule{}.Destroy(*value)
	}
}



type FfiConverterOptionalDecimal struct{}

var FfiConverterOptionalDecimalINSTANCE = FfiConverterOptionalDecimal{}

func (c FfiConverterOptionalDecimal) Lift(rb RustBufferI) **Decimal {
	return LiftFromRustBuffer[**Decimal](c, rb)
}

func (_ FfiConverterOptionalDecimal) Read(reader io.Reader) **Decimal {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterDecimalINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalDecimal) Lower(value **Decimal) RustBuffer {
	return LowerIntoRustBuffer[**Decimal](c, value)
}

func (_ FfiConverterOptionalDecimal) Write(writer io.Writer, value **Decimal) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterDecimalINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalDecimal struct {}

func (_ FfiDestroyerOptionalDecimal) Destroy(value **Decimal) {
	if value != nil {
		FfiDestroyerDecimal{}.Destroy(*value)
	}
}



type FfiConverterOptionalPreciseDecimal struct{}

var FfiConverterOptionalPreciseDecimalINSTANCE = FfiConverterOptionalPreciseDecimal{}

func (c FfiConverterOptionalPreciseDecimal) Lift(rb RustBufferI) **PreciseDecimal {
	return LiftFromRustBuffer[**PreciseDecimal](c, rb)
}

func (_ FfiConverterOptionalPreciseDecimal) Read(reader io.Reader) **PreciseDecimal {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterPreciseDecimalINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalPreciseDecimal) Lower(value **PreciseDecimal) RustBuffer {
	return LowerIntoRustBuffer[**PreciseDecimal](c, value)
}

func (_ FfiConverterOptionalPreciseDecimal) Write(writer io.Writer, value **PreciseDecimal) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterPreciseDecimalINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalPreciseDecimal struct {}

func (_ FfiDestroyerOptionalPreciseDecimal) Destroy(value **PreciseDecimal) {
	if value != nil {
		FfiDestroyerPreciseDecimal{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeLockFeeModification struct{}

var FfiConverterOptionalTypeLockFeeModificationINSTANCE = FfiConverterOptionalTypeLockFeeModification{}

func (c FfiConverterOptionalTypeLockFeeModification) Lift(rb RustBufferI) *LockFeeModification {
	return LiftFromRustBuffer[*LockFeeModification](c, rb)
}

func (_ FfiConverterOptionalTypeLockFeeModification) Read(reader io.Reader) *LockFeeModification {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeLockFeeModificationINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeLockFeeModification) Lower(value *LockFeeModification) RustBuffer {
	return LowerIntoRustBuffer[*LockFeeModification](c, value)
}

func (_ FfiConverterOptionalTypeLockFeeModification) Write(writer io.Writer, value *LockFeeModification) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeLockFeeModificationINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeLockFeeModification struct {}

func (_ FfiDestroyerOptionalTypeLockFeeModification) Destroy(value *LockFeeModification) {
	if value != nil {
		FfiDestroyerTypeLockFeeModification{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeManifestBuilderAddressReservation struct{}

var FfiConverterOptionalTypeManifestBuilderAddressReservationINSTANCE = FfiConverterOptionalTypeManifestBuilderAddressReservation{}

func (c FfiConverterOptionalTypeManifestBuilderAddressReservation) Lift(rb RustBufferI) *ManifestBuilderAddressReservation {
	return LiftFromRustBuffer[*ManifestBuilderAddressReservation](c, rb)
}

func (_ FfiConverterOptionalTypeManifestBuilderAddressReservation) Read(reader io.Reader) *ManifestBuilderAddressReservation {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeManifestBuilderAddressReservationINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeManifestBuilderAddressReservation) Lower(value *ManifestBuilderAddressReservation) RustBuffer {
	return LowerIntoRustBuffer[*ManifestBuilderAddressReservation](c, value)
}

func (_ FfiConverterOptionalTypeManifestBuilderAddressReservation) Write(writer io.Writer, value *ManifestBuilderAddressReservation) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeManifestBuilderAddressReservationINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeManifestBuilderAddressReservation struct {}

func (_ FfiDestroyerOptionalTypeManifestBuilderAddressReservation) Destroy(value *ManifestBuilderAddressReservation) {
	if value != nil {
		FfiDestroyerTypeManifestBuilderAddressReservation{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeResourceManagerRole struct{}

var FfiConverterOptionalTypeResourceManagerRoleINSTANCE = FfiConverterOptionalTypeResourceManagerRole{}

func (c FfiConverterOptionalTypeResourceManagerRole) Lift(rb RustBufferI) *ResourceManagerRole {
	return LiftFromRustBuffer[*ResourceManagerRole](c, rb)
}

func (_ FfiConverterOptionalTypeResourceManagerRole) Read(reader io.Reader) *ResourceManagerRole {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeResourceManagerRoleINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeResourceManagerRole) Lower(value *ResourceManagerRole) RustBuffer {
	return LowerIntoRustBuffer[*ResourceManagerRole](c, value)
}

func (_ FfiConverterOptionalTypeResourceManagerRole) Write(writer io.Writer, value *ResourceManagerRole) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeResourceManagerRoleINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeResourceManagerRole struct {}

func (_ FfiDestroyerOptionalTypeResourceManagerRole) Destroy(value *ResourceManagerRole) {
	if value != nil {
		FfiDestroyerTypeResourceManagerRole{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeSchema struct{}

var FfiConverterOptionalTypeSchemaINSTANCE = FfiConverterOptionalTypeSchema{}

func (c FfiConverterOptionalTypeSchema) Lift(rb RustBufferI) *Schema {
	return LiftFromRustBuffer[*Schema](c, rb)
}

func (_ FfiConverterOptionalTypeSchema) Read(reader io.Reader) *Schema {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeSchemaINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeSchema) Lower(value *Schema) RustBuffer {
	return LowerIntoRustBuffer[*Schema](c, value)
}

func (_ FfiConverterOptionalTypeSchema) Write(writer io.Writer, value *Schema) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeSchemaINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeSchema struct {}

func (_ FfiDestroyerOptionalTypeSchema) Destroy(value *Schema) {
	if value != nil {
		FfiDestroyerTypeSchema{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeMetadataValue struct{}

var FfiConverterOptionalTypeMetadataValueINSTANCE = FfiConverterOptionalTypeMetadataValue{}

func (c FfiConverterOptionalTypeMetadataValue) Lift(rb RustBufferI) *MetadataValue {
	return LiftFromRustBuffer[*MetadataValue](c, rb)
}

func (_ FfiConverterOptionalTypeMetadataValue) Read(reader io.Reader) *MetadataValue {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeMetadataValueINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeMetadataValue) Lower(value *MetadataValue) RustBuffer {
	return LowerIntoRustBuffer[*MetadataValue](c, value)
}

func (_ FfiConverterOptionalTypeMetadataValue) Write(writer io.Writer, value *MetadataValue) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeMetadataValueINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeMetadataValue struct {}

func (_ FfiDestroyerOptionalTypeMetadataValue) Destroy(value *MetadataValue) {
	if value != nil {
		FfiDestroyerTypeMetadataValue{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeResourceOrNonFungible struct{}

var FfiConverterOptionalTypeResourceOrNonFungibleINSTANCE = FfiConverterOptionalTypeResourceOrNonFungible{}

func (c FfiConverterOptionalTypeResourceOrNonFungible) Lift(rb RustBufferI) *ResourceOrNonFungible {
	return LiftFromRustBuffer[*ResourceOrNonFungible](c, rb)
}

func (_ FfiConverterOptionalTypeResourceOrNonFungible) Read(reader io.Reader) *ResourceOrNonFungible {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeResourceOrNonFungibleINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeResourceOrNonFungible) Lower(value *ResourceOrNonFungible) RustBuffer {
	return LowerIntoRustBuffer[*ResourceOrNonFungible](c, value)
}

func (_ FfiConverterOptionalTypeResourceOrNonFungible) Write(writer io.Writer, value *ResourceOrNonFungible) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeResourceOrNonFungibleINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeResourceOrNonFungible struct {}

func (_ FfiDestroyerOptionalTypeResourceOrNonFungible) Destroy(value *ResourceOrNonFungible) {
	if value != nil {
		FfiDestroyerTypeResourceOrNonFungible{}.Destroy(*value)
	}
}



type FfiConverterSequenceUint32 struct{}

var FfiConverterSequenceUint32INSTANCE = FfiConverterSequenceUint32{}

func (c FfiConverterSequenceUint32) Lift(rb RustBufferI) []uint32 {
	return LiftFromRustBuffer[[]uint32](c, rb)
}

func (c FfiConverterSequenceUint32) Read(reader io.Reader) []uint32 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]uint32, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterUint32INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceUint32) Lower(value []uint32) RustBuffer {
	return LowerIntoRustBuffer[[]uint32](c, value)
}

func (c FfiConverterSequenceUint32) Write(writer io.Writer, value []uint32) {
	if len(value) > math.MaxInt32 {
		panic("[]uint32 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterUint32INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceUint32 struct {}

func (FfiDestroyerSequenceUint32) Destroy(sequence []uint32) {
	for _, value := range sequence {
		FfiDestroyerUint32{}.Destroy(value)	
	}
}



type FfiConverterSequenceInt32 struct{}

var FfiConverterSequenceInt32INSTANCE = FfiConverterSequenceInt32{}

func (c FfiConverterSequenceInt32) Lift(rb RustBufferI) []int32 {
	return LiftFromRustBuffer[[]int32](c, rb)
}

func (c FfiConverterSequenceInt32) Read(reader io.Reader) []int32 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]int32, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterInt32INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceInt32) Lower(value []int32) RustBuffer {
	return LowerIntoRustBuffer[[]int32](c, value)
}

func (c FfiConverterSequenceInt32) Write(writer io.Writer, value []int32) {
	if len(value) > math.MaxInt32 {
		panic("[]int32 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterInt32INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceInt32 struct {}

func (FfiDestroyerSequenceInt32) Destroy(sequence []int32) {
	for _, value := range sequence {
		FfiDestroyerInt32{}.Destroy(value)	
	}
}



type FfiConverterSequenceUint64 struct{}

var FfiConverterSequenceUint64INSTANCE = FfiConverterSequenceUint64{}

func (c FfiConverterSequenceUint64) Lift(rb RustBufferI) []uint64 {
	return LiftFromRustBuffer[[]uint64](c, rb)
}

func (c FfiConverterSequenceUint64) Read(reader io.Reader) []uint64 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]uint64, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterUint64INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceUint64) Lower(value []uint64) RustBuffer {
	return LowerIntoRustBuffer[[]uint64](c, value)
}

func (c FfiConverterSequenceUint64) Write(writer io.Writer, value []uint64) {
	if len(value) > math.MaxInt32 {
		panic("[]uint64 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterUint64INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceUint64 struct {}

func (FfiDestroyerSequenceUint64) Destroy(sequence []uint64) {
	for _, value := range sequence {
		FfiDestroyerUint64{}.Destroy(value)	
	}
}



type FfiConverterSequenceInt64 struct{}

var FfiConverterSequenceInt64INSTANCE = FfiConverterSequenceInt64{}

func (c FfiConverterSequenceInt64) Lift(rb RustBufferI) []int64 {
	return LiftFromRustBuffer[[]int64](c, rb)
}

func (c FfiConverterSequenceInt64) Read(reader io.Reader) []int64 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]int64, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterInt64INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceInt64) Lower(value []int64) RustBuffer {
	return LowerIntoRustBuffer[[]int64](c, value)
}

func (c FfiConverterSequenceInt64) Write(writer io.Writer, value []int64) {
	if len(value) > math.MaxInt32 {
		panic("[]int64 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterInt64INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceInt64 struct {}

func (FfiDestroyerSequenceInt64) Destroy(sequence []int64) {
	for _, value := range sequence {
		FfiDestroyerInt64{}.Destroy(value)	
	}
}



type FfiConverterSequenceBool struct{}

var FfiConverterSequenceBoolINSTANCE = FfiConverterSequenceBool{}

func (c FfiConverterSequenceBool) Lift(rb RustBufferI) []bool {
	return LiftFromRustBuffer[[]bool](c, rb)
}

func (c FfiConverterSequenceBool) Read(reader io.Reader) []bool {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]bool, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterBoolINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceBool) Lower(value []bool) RustBuffer {
	return LowerIntoRustBuffer[[]bool](c, value)
}

func (c FfiConverterSequenceBool) Write(writer io.Writer, value []bool) {
	if len(value) > math.MaxInt32 {
		panic("[]bool is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterBoolINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceBool struct {}

func (FfiDestroyerSequenceBool) Destroy(sequence []bool) {
	for _, value := range sequence {
		FfiDestroyerBool{}.Destroy(value)	
	}
}



type FfiConverterSequenceString struct{}

var FfiConverterSequenceStringINSTANCE = FfiConverterSequenceString{}

func (c FfiConverterSequenceString) Lift(rb RustBufferI) []string {
	return LiftFromRustBuffer[[]string](c, rb)
}

func (c FfiConverterSequenceString) Read(reader io.Reader) []string {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]string, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterStringINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceString) Lower(value []string) RustBuffer {
	return LowerIntoRustBuffer[[]string](c, value)
}

func (c FfiConverterSequenceString) Write(writer io.Writer, value []string) {
	if len(value) > math.MaxInt32 {
		panic("[]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterStringINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceString struct {}

func (FfiDestroyerSequenceString) Destroy(sequence []string) {
	for _, value := range sequence {
		FfiDestroyerString{}.Destroy(value)	
	}
}



type FfiConverterSequenceBytes struct{}

var FfiConverterSequenceBytesINSTANCE = FfiConverterSequenceBytes{}

func (c FfiConverterSequenceBytes) Lift(rb RustBufferI) [][]byte {
	return LiftFromRustBuffer[[][]byte](c, rb)
}

func (c FfiConverterSequenceBytes) Read(reader io.Reader) [][]byte {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([][]byte, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterBytesINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceBytes) Lower(value [][]byte) RustBuffer {
	return LowerIntoRustBuffer[[][]byte](c, value)
}

func (c FfiConverterSequenceBytes) Write(writer io.Writer, value [][]byte) {
	if len(value) > math.MaxInt32 {
		panic("[][]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterBytesINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceBytes struct {}

func (FfiDestroyerSequenceBytes) Destroy(sequence [][]byte) {
	for _, value := range sequence {
		FfiDestroyerBytes{}.Destroy(value)	
	}
}



type FfiConverterSequenceAddress struct{}

var FfiConverterSequenceAddressINSTANCE = FfiConverterSequenceAddress{}

func (c FfiConverterSequenceAddress) Lift(rb RustBufferI) []*Address {
	return LiftFromRustBuffer[[]*Address](c, rb)
}

func (c FfiConverterSequenceAddress) Read(reader io.Reader) []*Address {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]*Address, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterAddressINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceAddress) Lower(value []*Address) RustBuffer {
	return LowerIntoRustBuffer[[]*Address](c, value)
}

func (c FfiConverterSequenceAddress) Write(writer io.Writer, value []*Address) {
	if len(value) > math.MaxInt32 {
		panic("[]*Address is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterAddressINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceAddress struct {}

func (FfiDestroyerSequenceAddress) Destroy(sequence []*Address) {
	for _, value := range sequence {
		FfiDestroyerAddress{}.Destroy(value)	
	}
}



type FfiConverterSequenceDecimal struct{}

var FfiConverterSequenceDecimalINSTANCE = FfiConverterSequenceDecimal{}

func (c FfiConverterSequenceDecimal) Lift(rb RustBufferI) []*Decimal {
	return LiftFromRustBuffer[[]*Decimal](c, rb)
}

func (c FfiConverterSequenceDecimal) Read(reader io.Reader) []*Decimal {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]*Decimal, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterDecimalINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceDecimal) Lower(value []*Decimal) RustBuffer {
	return LowerIntoRustBuffer[[]*Decimal](c, value)
}

func (c FfiConverterSequenceDecimal) Write(writer io.Writer, value []*Decimal) {
	if len(value) > math.MaxInt32 {
		panic("[]*Decimal is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterDecimalINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceDecimal struct {}

func (FfiDestroyerSequenceDecimal) Destroy(sequence []*Decimal) {
	for _, value := range sequence {
		FfiDestroyerDecimal{}.Destroy(value)	
	}
}



type FfiConverterSequenceNonFungibleGlobalId struct{}

var FfiConverterSequenceNonFungibleGlobalIdINSTANCE = FfiConverterSequenceNonFungibleGlobalId{}

func (c FfiConverterSequenceNonFungibleGlobalId) Lift(rb RustBufferI) []*NonFungibleGlobalId {
	return LiftFromRustBuffer[[]*NonFungibleGlobalId](c, rb)
}

func (c FfiConverterSequenceNonFungibleGlobalId) Read(reader io.Reader) []*NonFungibleGlobalId {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]*NonFungibleGlobalId, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterNonFungibleGlobalIdINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceNonFungibleGlobalId) Lower(value []*NonFungibleGlobalId) RustBuffer {
	return LowerIntoRustBuffer[[]*NonFungibleGlobalId](c, value)
}

func (c FfiConverterSequenceNonFungibleGlobalId) Write(writer io.Writer, value []*NonFungibleGlobalId) {
	if len(value) > math.MaxInt32 {
		panic("[]*NonFungibleGlobalId is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterNonFungibleGlobalIdINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceNonFungibleGlobalId struct {}

func (FfiDestroyerSequenceNonFungibleGlobalId) Destroy(sequence []*NonFungibleGlobalId) {
	for _, value := range sequence {
		FfiDestroyerNonFungibleGlobalId{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeIndexedAssertion struct{}

var FfiConverterSequenceTypeIndexedAssertionINSTANCE = FfiConverterSequenceTypeIndexedAssertion{}

func (c FfiConverterSequenceTypeIndexedAssertion) Lift(rb RustBufferI) []IndexedAssertion {
	return LiftFromRustBuffer[[]IndexedAssertion](c, rb)
}

func (c FfiConverterSequenceTypeIndexedAssertion) Read(reader io.Reader) []IndexedAssertion {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]IndexedAssertion, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeIndexedAssertionINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeIndexedAssertion) Lower(value []IndexedAssertion) RustBuffer {
	return LowerIntoRustBuffer[[]IndexedAssertion](c, value)
}

func (c FfiConverterSequenceTypeIndexedAssertion) Write(writer io.Writer, value []IndexedAssertion) {
	if len(value) > math.MaxInt32 {
		panic("[]IndexedAssertion is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeIndexedAssertionINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeIndexedAssertion struct {}

func (FfiDestroyerSequenceTypeIndexedAssertion) Destroy(sequence []IndexedAssertion) {
	for _, value := range sequence {
		FfiDestroyerTypeIndexedAssertion{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeManifestBuilderBucket struct{}

var FfiConverterSequenceTypeManifestBuilderBucketINSTANCE = FfiConverterSequenceTypeManifestBuilderBucket{}

func (c FfiConverterSequenceTypeManifestBuilderBucket) Lift(rb RustBufferI) []ManifestBuilderBucket {
	return LiftFromRustBuffer[[]ManifestBuilderBucket](c, rb)
}

func (c FfiConverterSequenceTypeManifestBuilderBucket) Read(reader io.Reader) []ManifestBuilderBucket {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ManifestBuilderBucket, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeManifestBuilderBucketINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeManifestBuilderBucket) Lower(value []ManifestBuilderBucket) RustBuffer {
	return LowerIntoRustBuffer[[]ManifestBuilderBucket](c, value)
}

func (c FfiConverterSequenceTypeManifestBuilderBucket) Write(writer io.Writer, value []ManifestBuilderBucket) {
	if len(value) > math.MaxInt32 {
		panic("[]ManifestBuilderBucket is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeManifestBuilderBucketINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeManifestBuilderBucket struct {}

func (FfiDestroyerSequenceTypeManifestBuilderBucket) Destroy(sequence []ManifestBuilderBucket) {
	for _, value := range sequence {
		FfiDestroyerTypeManifestBuilderBucket{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeManifestBuilderMapEntry struct{}

var FfiConverterSequenceTypeManifestBuilderMapEntryINSTANCE = FfiConverterSequenceTypeManifestBuilderMapEntry{}

func (c FfiConverterSequenceTypeManifestBuilderMapEntry) Lift(rb RustBufferI) []ManifestBuilderMapEntry {
	return LiftFromRustBuffer[[]ManifestBuilderMapEntry](c, rb)
}

func (c FfiConverterSequenceTypeManifestBuilderMapEntry) Read(reader io.Reader) []ManifestBuilderMapEntry {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ManifestBuilderMapEntry, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeManifestBuilderMapEntryINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeManifestBuilderMapEntry) Lower(value []ManifestBuilderMapEntry) RustBuffer {
	return LowerIntoRustBuffer[[]ManifestBuilderMapEntry](c, value)
}

func (c FfiConverterSequenceTypeManifestBuilderMapEntry) Write(writer io.Writer, value []ManifestBuilderMapEntry) {
	if len(value) > math.MaxInt32 {
		panic("[]ManifestBuilderMapEntry is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeManifestBuilderMapEntryINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeManifestBuilderMapEntry struct {}

func (FfiDestroyerSequenceTypeManifestBuilderMapEntry) Destroy(sequence []ManifestBuilderMapEntry) {
	for _, value := range sequence {
		FfiDestroyerTypeManifestBuilderMapEntry{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeMapEntry struct{}

var FfiConverterSequenceTypeMapEntryINSTANCE = FfiConverterSequenceTypeMapEntry{}

func (c FfiConverterSequenceTypeMapEntry) Lift(rb RustBufferI) []MapEntry {
	return LiftFromRustBuffer[[]MapEntry](c, rb)
}

func (c FfiConverterSequenceTypeMapEntry) Read(reader io.Reader) []MapEntry {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]MapEntry, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeMapEntryINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeMapEntry) Lower(value []MapEntry) RustBuffer {
	return LowerIntoRustBuffer[[]MapEntry](c, value)
}

func (c FfiConverterSequenceTypeMapEntry) Write(writer io.Writer, value []MapEntry) {
	if len(value) > math.MaxInt32 {
		panic("[]MapEntry is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeMapEntryINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeMapEntry struct {}

func (FfiDestroyerSequenceTypeMapEntry) Destroy(sequence []MapEntry) {
	for _, value := range sequence {
		FfiDestroyerTypeMapEntry{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTrackedPoolContribution struct{}

var FfiConverterSequenceTypeTrackedPoolContributionINSTANCE = FfiConverterSequenceTypeTrackedPoolContribution{}

func (c FfiConverterSequenceTypeTrackedPoolContribution) Lift(rb RustBufferI) []TrackedPoolContribution {
	return LiftFromRustBuffer[[]TrackedPoolContribution](c, rb)
}

func (c FfiConverterSequenceTypeTrackedPoolContribution) Read(reader io.Reader) []TrackedPoolContribution {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TrackedPoolContribution, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTrackedPoolContributionINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTrackedPoolContribution) Lower(value []TrackedPoolContribution) RustBuffer {
	return LowerIntoRustBuffer[[]TrackedPoolContribution](c, value)
}

func (c FfiConverterSequenceTypeTrackedPoolContribution) Write(writer io.Writer, value []TrackedPoolContribution) {
	if len(value) > math.MaxInt32 {
		panic("[]TrackedPoolContribution is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTrackedPoolContributionINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTrackedPoolContribution struct {}

func (FfiDestroyerSequenceTypeTrackedPoolContribution) Destroy(sequence []TrackedPoolContribution) {
	for _, value := range sequence {
		FfiDestroyerTypeTrackedPoolContribution{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTrackedPoolRedemption struct{}

var FfiConverterSequenceTypeTrackedPoolRedemptionINSTANCE = FfiConverterSequenceTypeTrackedPoolRedemption{}

func (c FfiConverterSequenceTypeTrackedPoolRedemption) Lift(rb RustBufferI) []TrackedPoolRedemption {
	return LiftFromRustBuffer[[]TrackedPoolRedemption](c, rb)
}

func (c FfiConverterSequenceTypeTrackedPoolRedemption) Read(reader io.Reader) []TrackedPoolRedemption {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TrackedPoolRedemption, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTrackedPoolRedemptionINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTrackedPoolRedemption) Lower(value []TrackedPoolRedemption) RustBuffer {
	return LowerIntoRustBuffer[[]TrackedPoolRedemption](c, value)
}

func (c FfiConverterSequenceTypeTrackedPoolRedemption) Write(writer io.Writer, value []TrackedPoolRedemption) {
	if len(value) > math.MaxInt32 {
		panic("[]TrackedPoolRedemption is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTrackedPoolRedemptionINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTrackedPoolRedemption struct {}

func (FfiDestroyerSequenceTypeTrackedPoolRedemption) Destroy(sequence []TrackedPoolRedemption) {
	for _, value := range sequence {
		FfiDestroyerTypeTrackedPoolRedemption{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTrackedValidatorClaim struct{}

var FfiConverterSequenceTypeTrackedValidatorClaimINSTANCE = FfiConverterSequenceTypeTrackedValidatorClaim{}

func (c FfiConverterSequenceTypeTrackedValidatorClaim) Lift(rb RustBufferI) []TrackedValidatorClaim {
	return LiftFromRustBuffer[[]TrackedValidatorClaim](c, rb)
}

func (c FfiConverterSequenceTypeTrackedValidatorClaim) Read(reader io.Reader) []TrackedValidatorClaim {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TrackedValidatorClaim, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTrackedValidatorClaimINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTrackedValidatorClaim) Lower(value []TrackedValidatorClaim) RustBuffer {
	return LowerIntoRustBuffer[[]TrackedValidatorClaim](c, value)
}

func (c FfiConverterSequenceTypeTrackedValidatorClaim) Write(writer io.Writer, value []TrackedValidatorClaim) {
	if len(value) > math.MaxInt32 {
		panic("[]TrackedValidatorClaim is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTrackedValidatorClaimINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTrackedValidatorClaim struct {}

func (FfiDestroyerSequenceTypeTrackedValidatorClaim) Destroy(sequence []TrackedValidatorClaim) {
	for _, value := range sequence {
		FfiDestroyerTypeTrackedValidatorClaim{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTrackedValidatorStake struct{}

var FfiConverterSequenceTypeTrackedValidatorStakeINSTANCE = FfiConverterSequenceTypeTrackedValidatorStake{}

func (c FfiConverterSequenceTypeTrackedValidatorStake) Lift(rb RustBufferI) []TrackedValidatorStake {
	return LiftFromRustBuffer[[]TrackedValidatorStake](c, rb)
}

func (c FfiConverterSequenceTypeTrackedValidatorStake) Read(reader io.Reader) []TrackedValidatorStake {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TrackedValidatorStake, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTrackedValidatorStakeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTrackedValidatorStake) Lower(value []TrackedValidatorStake) RustBuffer {
	return LowerIntoRustBuffer[[]TrackedValidatorStake](c, value)
}

func (c FfiConverterSequenceTypeTrackedValidatorStake) Write(writer io.Writer, value []TrackedValidatorStake) {
	if len(value) > math.MaxInt32 {
		panic("[]TrackedValidatorStake is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTrackedValidatorStakeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTrackedValidatorStake struct {}

func (FfiDestroyerSequenceTypeTrackedValidatorStake) Destroy(sequence []TrackedValidatorStake) {
	for _, value := range sequence {
		FfiDestroyerTypeTrackedValidatorStake{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTrackedValidatorUnstake struct{}

var FfiConverterSequenceTypeTrackedValidatorUnstakeINSTANCE = FfiConverterSequenceTypeTrackedValidatorUnstake{}

func (c FfiConverterSequenceTypeTrackedValidatorUnstake) Lift(rb RustBufferI) []TrackedValidatorUnstake {
	return LiftFromRustBuffer[[]TrackedValidatorUnstake](c, rb)
}

func (c FfiConverterSequenceTypeTrackedValidatorUnstake) Read(reader io.Reader) []TrackedValidatorUnstake {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TrackedValidatorUnstake, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTrackedValidatorUnstakeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTrackedValidatorUnstake) Lower(value []TrackedValidatorUnstake) RustBuffer {
	return LowerIntoRustBuffer[[]TrackedValidatorUnstake](c, value)
}

func (c FfiConverterSequenceTypeTrackedValidatorUnstake) Write(writer io.Writer, value []TrackedValidatorUnstake) {
	if len(value) > math.MaxInt32 {
		panic("[]TrackedValidatorUnstake is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTrackedValidatorUnstakeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTrackedValidatorUnstake struct {}

func (FfiDestroyerSequenceTypeTrackedValidatorUnstake) Destroy(sequence []TrackedValidatorUnstake) {
	for _, value := range sequence {
		FfiDestroyerTypeTrackedValidatorUnstake{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeUnstakeDataEntry struct{}

var FfiConverterSequenceTypeUnstakeDataEntryINSTANCE = FfiConverterSequenceTypeUnstakeDataEntry{}

func (c FfiConverterSequenceTypeUnstakeDataEntry) Lift(rb RustBufferI) []UnstakeDataEntry {
	return LiftFromRustBuffer[[]UnstakeDataEntry](c, rb)
}

func (c FfiConverterSequenceTypeUnstakeDataEntry) Read(reader io.Reader) []UnstakeDataEntry {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]UnstakeDataEntry, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeUnstakeDataEntryINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeUnstakeDataEntry) Lower(value []UnstakeDataEntry) RustBuffer {
	return LowerIntoRustBuffer[[]UnstakeDataEntry](c, value)
}

func (c FfiConverterSequenceTypeUnstakeDataEntry) Write(writer io.Writer, value []UnstakeDataEntry) {
	if len(value) > math.MaxInt32 {
		panic("[]UnstakeDataEntry is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeUnstakeDataEntryINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeUnstakeDataEntry struct {}

func (FfiDestroyerSequenceTypeUnstakeDataEntry) Destroy(sequence []UnstakeDataEntry) {
	for _, value := range sequence {
		FfiDestroyerTypeUnstakeDataEntry{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeDetailedManifestClass struct{}

var FfiConverterSequenceTypeDetailedManifestClassINSTANCE = FfiConverterSequenceTypeDetailedManifestClass{}

func (c FfiConverterSequenceTypeDetailedManifestClass) Lift(rb RustBufferI) []DetailedManifestClass {
	return LiftFromRustBuffer[[]DetailedManifestClass](c, rb)
}

func (c FfiConverterSequenceTypeDetailedManifestClass) Read(reader io.Reader) []DetailedManifestClass {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]DetailedManifestClass, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeDetailedManifestClassINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeDetailedManifestClass) Lower(value []DetailedManifestClass) RustBuffer {
	return LowerIntoRustBuffer[[]DetailedManifestClass](c, value)
}

func (c FfiConverterSequenceTypeDetailedManifestClass) Write(writer io.Writer, value []DetailedManifestClass) {
	if len(value) > math.MaxInt32 {
		panic("[]DetailedManifestClass is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeDetailedManifestClassINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeDetailedManifestClass struct {}

func (FfiDestroyerSequenceTypeDetailedManifestClass) Destroy(sequence []DetailedManifestClass) {
	for _, value := range sequence {
		FfiDestroyerTypeDetailedManifestClass{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeEntityType struct{}

var FfiConverterSequenceTypeEntityTypeINSTANCE = FfiConverterSequenceTypeEntityType{}

func (c FfiConverterSequenceTypeEntityType) Lift(rb RustBufferI) []EntityType {
	return LiftFromRustBuffer[[]EntityType](c, rb)
}

func (c FfiConverterSequenceTypeEntityType) Read(reader io.Reader) []EntityType {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]EntityType, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeEntityTypeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeEntityType) Lower(value []EntityType) RustBuffer {
	return LowerIntoRustBuffer[[]EntityType](c, value)
}

func (c FfiConverterSequenceTypeEntityType) Write(writer io.Writer, value []EntityType) {
	if len(value) > math.MaxInt32 {
		panic("[]EntityType is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeEntityTypeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeEntityType struct {}

func (FfiDestroyerSequenceTypeEntityType) Destroy(sequence []EntityType) {
	for _, value := range sequence {
		FfiDestroyerTypeEntityType{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeInstruction struct{}

var FfiConverterSequenceTypeInstructionINSTANCE = FfiConverterSequenceTypeInstruction{}

func (c FfiConverterSequenceTypeInstruction) Lift(rb RustBufferI) []Instruction {
	return LiftFromRustBuffer[[]Instruction](c, rb)
}

func (c FfiConverterSequenceTypeInstruction) Read(reader io.Reader) []Instruction {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Instruction, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeInstructionINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeInstruction) Lower(value []Instruction) RustBuffer {
	return LowerIntoRustBuffer[[]Instruction](c, value)
}

func (c FfiConverterSequenceTypeInstruction) Write(writer io.Writer, value []Instruction) {
	if len(value) > math.MaxInt32 {
		panic("[]Instruction is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeInstructionINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeInstruction struct {}

func (FfiDestroyerSequenceTypeInstruction) Destroy(sequence []Instruction) {
	for _, value := range sequence {
		FfiDestroyerTypeInstruction{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeManifestBuilderValue struct{}

var FfiConverterSequenceTypeManifestBuilderValueINSTANCE = FfiConverterSequenceTypeManifestBuilderValue{}

func (c FfiConverterSequenceTypeManifestBuilderValue) Lift(rb RustBufferI) []ManifestBuilderValue {
	return LiftFromRustBuffer[[]ManifestBuilderValue](c, rb)
}

func (c FfiConverterSequenceTypeManifestBuilderValue) Read(reader io.Reader) []ManifestBuilderValue {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ManifestBuilderValue, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeManifestBuilderValueINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeManifestBuilderValue) Lower(value []ManifestBuilderValue) RustBuffer {
	return LowerIntoRustBuffer[[]ManifestBuilderValue](c, value)
}

func (c FfiConverterSequenceTypeManifestBuilderValue) Write(writer io.Writer, value []ManifestBuilderValue) {
	if len(value) > math.MaxInt32 {
		panic("[]ManifestBuilderValue is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeManifestBuilderValueINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeManifestBuilderValue struct {}

func (FfiDestroyerSequenceTypeManifestBuilderValue) Destroy(sequence []ManifestBuilderValue) {
	for _, value := range sequence {
		FfiDestroyerTypeManifestBuilderValue{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeManifestClass struct{}

var FfiConverterSequenceTypeManifestClassINSTANCE = FfiConverterSequenceTypeManifestClass{}

func (c FfiConverterSequenceTypeManifestClass) Lift(rb RustBufferI) []ManifestClass {
	return LiftFromRustBuffer[[]ManifestClass](c, rb)
}

func (c FfiConverterSequenceTypeManifestClass) Read(reader io.Reader) []ManifestClass {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ManifestClass, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeManifestClassINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeManifestClass) Lower(value []ManifestClass) RustBuffer {
	return LowerIntoRustBuffer[[]ManifestClass](c, value)
}

func (c FfiConverterSequenceTypeManifestClass) Write(writer io.Writer, value []ManifestClass) {
	if len(value) > math.MaxInt32 {
		panic("[]ManifestClass is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeManifestClassINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeManifestClass struct {}

func (FfiDestroyerSequenceTypeManifestClass) Destroy(sequence []ManifestClass) {
	for _, value := range sequence {
		FfiDestroyerTypeManifestClass{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeManifestValue struct{}

var FfiConverterSequenceTypeManifestValueINSTANCE = FfiConverterSequenceTypeManifestValue{}

func (c FfiConverterSequenceTypeManifestValue) Lift(rb RustBufferI) []ManifestValue {
	return LiftFromRustBuffer[[]ManifestValue](c, rb)
}

func (c FfiConverterSequenceTypeManifestValue) Read(reader io.Reader) []ManifestValue {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ManifestValue, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeManifestValueINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeManifestValue) Lower(value []ManifestValue) RustBuffer {
	return LowerIntoRustBuffer[[]ManifestValue](c, value)
}

func (c FfiConverterSequenceTypeManifestValue) Write(writer io.Writer, value []ManifestValue) {
	if len(value) > math.MaxInt32 {
		panic("[]ManifestValue is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeManifestValueINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeManifestValue struct {}

func (FfiDestroyerSequenceTypeManifestValue) Destroy(sequence []ManifestValue) {
	for _, value := range sequence {
		FfiDestroyerTypeManifestValue{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeNonFungibleLocalId struct{}

var FfiConverterSequenceTypeNonFungibleLocalIdINSTANCE = FfiConverterSequenceTypeNonFungibleLocalId{}

func (c FfiConverterSequenceTypeNonFungibleLocalId) Lift(rb RustBufferI) []NonFungibleLocalId {
	return LiftFromRustBuffer[[]NonFungibleLocalId](c, rb)
}

func (c FfiConverterSequenceTypeNonFungibleLocalId) Read(reader io.Reader) []NonFungibleLocalId {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]NonFungibleLocalId, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeNonFungibleLocalIdINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeNonFungibleLocalId) Lower(value []NonFungibleLocalId) RustBuffer {
	return LowerIntoRustBuffer[[]NonFungibleLocalId](c, value)
}

func (c FfiConverterSequenceTypeNonFungibleLocalId) Write(writer io.Writer, value []NonFungibleLocalId) {
	if len(value) > math.MaxInt32 {
		panic("[]NonFungibleLocalId is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeNonFungibleLocalIdINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeNonFungibleLocalId struct {}

func (FfiDestroyerSequenceTypeNonFungibleLocalId) Destroy(sequence []NonFungibleLocalId) {
	for _, value := range sequence {
		FfiDestroyerTypeNonFungibleLocalId{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypePublicKey struct{}

var FfiConverterSequenceTypePublicKeyINSTANCE = FfiConverterSequenceTypePublicKey{}

func (c FfiConverterSequenceTypePublicKey) Lift(rb RustBufferI) []PublicKey {
	return LiftFromRustBuffer[[]PublicKey](c, rb)
}

func (c FfiConverterSequenceTypePublicKey) Read(reader io.Reader) []PublicKey {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PublicKey, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePublicKeyINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePublicKey) Lower(value []PublicKey) RustBuffer {
	return LowerIntoRustBuffer[[]PublicKey](c, value)
}

func (c FfiConverterSequenceTypePublicKey) Write(writer io.Writer, value []PublicKey) {
	if len(value) > math.MaxInt32 {
		panic("[]PublicKey is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePublicKeyINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePublicKey struct {}

func (FfiDestroyerSequenceTypePublicKey) Destroy(sequence []PublicKey) {
	for _, value := range sequence {
		FfiDestroyerTypePublicKey{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypePublicKeyHash struct{}

var FfiConverterSequenceTypePublicKeyHashINSTANCE = FfiConverterSequenceTypePublicKeyHash{}

func (c FfiConverterSequenceTypePublicKeyHash) Lift(rb RustBufferI) []PublicKeyHash {
	return LiftFromRustBuffer[[]PublicKeyHash](c, rb)
}

func (c FfiConverterSequenceTypePublicKeyHash) Read(reader io.Reader) []PublicKeyHash {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PublicKeyHash, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePublicKeyHashINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePublicKeyHash) Lower(value []PublicKeyHash) RustBuffer {
	return LowerIntoRustBuffer[[]PublicKeyHash](c, value)
}

func (c FfiConverterSequenceTypePublicKeyHash) Write(writer io.Writer, value []PublicKeyHash) {
	if len(value) > math.MaxInt32 {
		panic("[]PublicKeyHash is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePublicKeyHashINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePublicKeyHash struct {}

func (FfiDestroyerSequenceTypePublicKeyHash) Destroy(sequence []PublicKeyHash) {
	for _, value := range sequence {
		FfiDestroyerTypePublicKeyHash{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeReservedInstruction struct{}

var FfiConverterSequenceTypeReservedInstructionINSTANCE = FfiConverterSequenceTypeReservedInstruction{}

func (c FfiConverterSequenceTypeReservedInstruction) Lift(rb RustBufferI) []ReservedInstruction {
	return LiftFromRustBuffer[[]ReservedInstruction](c, rb)
}

func (c FfiConverterSequenceTypeReservedInstruction) Read(reader io.Reader) []ReservedInstruction {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ReservedInstruction, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeReservedInstructionINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeReservedInstruction) Lower(value []ReservedInstruction) RustBuffer {
	return LowerIntoRustBuffer[[]ReservedInstruction](c, value)
}

func (c FfiConverterSequenceTypeReservedInstruction) Write(writer io.Writer, value []ReservedInstruction) {
	if len(value) > math.MaxInt32 {
		panic("[]ReservedInstruction is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeReservedInstructionINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeReservedInstruction struct {}

func (FfiDestroyerSequenceTypeReservedInstruction) Destroy(sequence []ReservedInstruction) {
	for _, value := range sequence {
		FfiDestroyerTypeReservedInstruction{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeResourceIndicator struct{}

var FfiConverterSequenceTypeResourceIndicatorINSTANCE = FfiConverterSequenceTypeResourceIndicator{}

func (c FfiConverterSequenceTypeResourceIndicator) Lift(rb RustBufferI) []ResourceIndicator {
	return LiftFromRustBuffer[[]ResourceIndicator](c, rb)
}

func (c FfiConverterSequenceTypeResourceIndicator) Read(reader io.Reader) []ResourceIndicator {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ResourceIndicator, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeResourceIndicatorINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeResourceIndicator) Lower(value []ResourceIndicator) RustBuffer {
	return LowerIntoRustBuffer[[]ResourceIndicator](c, value)
}

func (c FfiConverterSequenceTypeResourceIndicator) Write(writer io.Writer, value []ResourceIndicator) {
	if len(value) > math.MaxInt32 {
		panic("[]ResourceIndicator is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeResourceIndicatorINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeResourceIndicator struct {}

func (FfiDestroyerSequenceTypeResourceIndicator) Destroy(sequence []ResourceIndicator) {
	for _, value := range sequence {
		FfiDestroyerTypeResourceIndicator{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeResourceOrNonFungible struct{}

var FfiConverterSequenceTypeResourceOrNonFungibleINSTANCE = FfiConverterSequenceTypeResourceOrNonFungible{}

func (c FfiConverterSequenceTypeResourceOrNonFungible) Lift(rb RustBufferI) []ResourceOrNonFungible {
	return LiftFromRustBuffer[[]ResourceOrNonFungible](c, rb)
}

func (c FfiConverterSequenceTypeResourceOrNonFungible) Read(reader io.Reader) []ResourceOrNonFungible {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ResourceOrNonFungible, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeResourceOrNonFungibleINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeResourceOrNonFungible) Lower(value []ResourceOrNonFungible) RustBuffer {
	return LowerIntoRustBuffer[[]ResourceOrNonFungible](c, value)
}

func (c FfiConverterSequenceTypeResourceOrNonFungible) Write(writer io.Writer, value []ResourceOrNonFungible) {
	if len(value) > math.MaxInt32 {
		panic("[]ResourceOrNonFungible is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeResourceOrNonFungibleINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeResourceOrNonFungible struct {}

func (FfiDestroyerSequenceTypeResourceOrNonFungible) Destroy(sequence []ResourceOrNonFungible) {
	for _, value := range sequence {
		FfiDestroyerTypeResourceOrNonFungible{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeResourceSpecifier struct{}

var FfiConverterSequenceTypeResourceSpecifierINSTANCE = FfiConverterSequenceTypeResourceSpecifier{}

func (c FfiConverterSequenceTypeResourceSpecifier) Lift(rb RustBufferI) []ResourceSpecifier {
	return LiftFromRustBuffer[[]ResourceSpecifier](c, rb)
}

func (c FfiConverterSequenceTypeResourceSpecifier) Read(reader io.Reader) []ResourceSpecifier {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ResourceSpecifier, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeResourceSpecifierINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeResourceSpecifier) Lower(value []ResourceSpecifier) RustBuffer {
	return LowerIntoRustBuffer[[]ResourceSpecifier](c, value)
}

func (c FfiConverterSequenceTypeResourceSpecifier) Write(writer io.Writer, value []ResourceSpecifier) {
	if len(value) > math.MaxInt32 {
		panic("[]ResourceSpecifier is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeResourceSpecifierINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeResourceSpecifier struct {}

func (FfiDestroyerSequenceTypeResourceSpecifier) Destroy(sequence []ResourceSpecifier) {
	for _, value := range sequence {
		FfiDestroyerTypeResourceSpecifier{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeSignatureWithPublicKey struct{}

var FfiConverterSequenceTypeSignatureWithPublicKeyINSTANCE = FfiConverterSequenceTypeSignatureWithPublicKey{}

func (c FfiConverterSequenceTypeSignatureWithPublicKey) Lift(rb RustBufferI) []SignatureWithPublicKey {
	return LiftFromRustBuffer[[]SignatureWithPublicKey](c, rb)
}

func (c FfiConverterSequenceTypeSignatureWithPublicKey) Read(reader io.Reader) []SignatureWithPublicKey {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]SignatureWithPublicKey, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeSignatureWithPublicKeyINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeSignatureWithPublicKey) Lower(value []SignatureWithPublicKey) RustBuffer {
	return LowerIntoRustBuffer[[]SignatureWithPublicKey](c, value)
}

func (c FfiConverterSequenceTypeSignatureWithPublicKey) Write(writer io.Writer, value []SignatureWithPublicKey) {
	if len(value) > math.MaxInt32 {
		panic("[]SignatureWithPublicKey is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeSignatureWithPublicKeyINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeSignatureWithPublicKey struct {}

func (FfiDestroyerSequenceTypeSignatureWithPublicKey) Destroy(sequence []SignatureWithPublicKey) {
	for _, value := range sequence {
		FfiDestroyerTypeSignatureWithPublicKey{}.Destroy(value)	
	}
}



type FfiConverterMapStringDecimal struct {}

var FfiConverterMapStringDecimalINSTANCE = FfiConverterMapStringDecimal{}

func (c FfiConverterMapStringDecimal) Lift(rb RustBufferI) map[string]*Decimal {
	return LiftFromRustBuffer[map[string]*Decimal](c, rb)
}

func (_ FfiConverterMapStringDecimal) Read(reader io.Reader) map[string]*Decimal {
	result := make(map[string]*Decimal)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterDecimalINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringDecimal) Lower(value map[string]*Decimal) RustBuffer {
	return LowerIntoRustBuffer[map[string]*Decimal](c, value)
}

func (_ FfiConverterMapStringDecimal) Write(writer io.Writer, mapValue map[string]*Decimal) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]*Decimal is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterDecimalINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringDecimal struct {}

func (_ FfiDestroyerMapStringDecimal) Destroy(mapValue map[string]*Decimal) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerDecimal{}.Destroy(value)	
	}
}



type FfiConverterMapStringTypeMetadataInitEntry struct {}

var FfiConverterMapStringTypeMetadataInitEntryINSTANCE = FfiConverterMapStringTypeMetadataInitEntry{}

func (c FfiConverterMapStringTypeMetadataInitEntry) Lift(rb RustBufferI) map[string]MetadataInitEntry {
	return LiftFromRustBuffer[map[string]MetadataInitEntry](c, rb)
}

func (_ FfiConverterMapStringTypeMetadataInitEntry) Read(reader io.Reader) map[string]MetadataInitEntry {
	result := make(map[string]MetadataInitEntry)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterTypeMetadataInitEntryINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringTypeMetadataInitEntry) Lower(value map[string]MetadataInitEntry) RustBuffer {
	return LowerIntoRustBuffer[map[string]MetadataInitEntry](c, value)
}

func (_ FfiConverterMapStringTypeMetadataInitEntry) Write(writer io.Writer, mapValue map[string]MetadataInitEntry) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]MetadataInitEntry is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterTypeMetadataInitEntryINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringTypeMetadataInitEntry struct {}

func (_ FfiDestroyerMapStringTypeMetadataInitEntry) Destroy(mapValue map[string]MetadataInitEntry) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerTypeMetadataInitEntry{}.Destroy(value)	
	}
}



type FfiConverterMapStringTypeValidatorInfo struct {}

var FfiConverterMapStringTypeValidatorInfoINSTANCE = FfiConverterMapStringTypeValidatorInfo{}

func (c FfiConverterMapStringTypeValidatorInfo) Lift(rb RustBufferI) map[string]ValidatorInfo {
	return LiftFromRustBuffer[map[string]ValidatorInfo](c, rb)
}

func (_ FfiConverterMapStringTypeValidatorInfo) Read(reader io.Reader) map[string]ValidatorInfo {
	result := make(map[string]ValidatorInfo)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterTypeValidatorInfoINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringTypeValidatorInfo) Lower(value map[string]ValidatorInfo) RustBuffer {
	return LowerIntoRustBuffer[map[string]ValidatorInfo](c, value)
}

func (_ FfiConverterMapStringTypeValidatorInfo) Write(writer io.Writer, mapValue map[string]ValidatorInfo) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]ValidatorInfo is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterTypeValidatorInfoINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringTypeValidatorInfo struct {}

func (_ FfiDestroyerMapStringTypeValidatorInfo) Destroy(mapValue map[string]ValidatorInfo) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerTypeValidatorInfo{}.Destroy(value)	
	}
}



type FfiConverterMapStringTypeAccountDefaultDepositRule struct {}

var FfiConverterMapStringTypeAccountDefaultDepositRuleINSTANCE = FfiConverterMapStringTypeAccountDefaultDepositRule{}

func (c FfiConverterMapStringTypeAccountDefaultDepositRule) Lift(rb RustBufferI) map[string]AccountDefaultDepositRule {
	return LiftFromRustBuffer[map[string]AccountDefaultDepositRule](c, rb)
}

func (_ FfiConverterMapStringTypeAccountDefaultDepositRule) Read(reader io.Reader) map[string]AccountDefaultDepositRule {
	result := make(map[string]AccountDefaultDepositRule)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterTypeAccountDefaultDepositRuleINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringTypeAccountDefaultDepositRule) Lower(value map[string]AccountDefaultDepositRule) RustBuffer {
	return LowerIntoRustBuffer[map[string]AccountDefaultDepositRule](c, value)
}

func (_ FfiConverterMapStringTypeAccountDefaultDepositRule) Write(writer io.Writer, mapValue map[string]AccountDefaultDepositRule) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]AccountDefaultDepositRule is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterTypeAccountDefaultDepositRuleINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringTypeAccountDefaultDepositRule struct {}

func (_ FfiDestroyerMapStringTypeAccountDefaultDepositRule) Destroy(mapValue map[string]AccountDefaultDepositRule) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerTypeAccountDefaultDepositRule{}.Destroy(value)	
	}
}



type FfiConverterMapStringTypeResourcePreferenceUpdate struct {}

var FfiConverterMapStringTypeResourcePreferenceUpdateINSTANCE = FfiConverterMapStringTypeResourcePreferenceUpdate{}

func (c FfiConverterMapStringTypeResourcePreferenceUpdate) Lift(rb RustBufferI) map[string]ResourcePreferenceUpdate {
	return LiftFromRustBuffer[map[string]ResourcePreferenceUpdate](c, rb)
}

func (_ FfiConverterMapStringTypeResourcePreferenceUpdate) Read(reader io.Reader) map[string]ResourcePreferenceUpdate {
	result := make(map[string]ResourcePreferenceUpdate)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterTypeResourcePreferenceUpdateINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringTypeResourcePreferenceUpdate) Lower(value map[string]ResourcePreferenceUpdate) RustBuffer {
	return LowerIntoRustBuffer[map[string]ResourcePreferenceUpdate](c, value)
}

func (_ FfiConverterMapStringTypeResourcePreferenceUpdate) Write(writer io.Writer, mapValue map[string]ResourcePreferenceUpdate) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]ResourcePreferenceUpdate is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterTypeResourcePreferenceUpdateINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringTypeResourcePreferenceUpdate struct {}

func (_ FfiDestroyerMapStringTypeResourcePreferenceUpdate) Destroy(mapValue map[string]ResourcePreferenceUpdate) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerTypeResourcePreferenceUpdate{}.Destroy(value)	
	}
}



type FfiConverterMapStringTypeResourceSpecifier struct {}

var FfiConverterMapStringTypeResourceSpecifierINSTANCE = FfiConverterMapStringTypeResourceSpecifier{}

func (c FfiConverterMapStringTypeResourceSpecifier) Lift(rb RustBufferI) map[string]ResourceSpecifier {
	return LiftFromRustBuffer[map[string]ResourceSpecifier](c, rb)
}

func (_ FfiConverterMapStringTypeResourceSpecifier) Read(reader io.Reader) map[string]ResourceSpecifier {
	result := make(map[string]ResourceSpecifier)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterTypeResourceSpecifierINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringTypeResourceSpecifier) Lower(value map[string]ResourceSpecifier) RustBuffer {
	return LowerIntoRustBuffer[map[string]ResourceSpecifier](c, value)
}

func (_ FfiConverterMapStringTypeResourceSpecifier) Write(writer io.Writer, mapValue map[string]ResourceSpecifier) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]ResourceSpecifier is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterTypeResourceSpecifierINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringTypeResourceSpecifier struct {}

func (_ FfiDestroyerMapStringTypeResourceSpecifier) Destroy(mapValue map[string]ResourceSpecifier) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerTypeResourceSpecifier{}.Destroy(value)	
	}
}



type FfiConverterMapStringOptionalAccessRule struct {}

var FfiConverterMapStringOptionalAccessRuleINSTANCE = FfiConverterMapStringOptionalAccessRule{}

func (c FfiConverterMapStringOptionalAccessRule) Lift(rb RustBufferI) map[string]**AccessRule {
	return LiftFromRustBuffer[map[string]**AccessRule](c, rb)
}

func (_ FfiConverterMapStringOptionalAccessRule) Read(reader io.Reader) map[string]**AccessRule {
	result := make(map[string]**AccessRule)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterOptionalAccessRuleINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringOptionalAccessRule) Lower(value map[string]**AccessRule) RustBuffer {
	return LowerIntoRustBuffer[map[string]**AccessRule](c, value)
}

func (_ FfiConverterMapStringOptionalAccessRule) Write(writer io.Writer, mapValue map[string]**AccessRule) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]**AccessRule is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterOptionalAccessRuleINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringOptionalAccessRule struct {}

func (_ FfiDestroyerMapStringOptionalAccessRule) Destroy(mapValue map[string]**AccessRule) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerOptionalAccessRule{}.Destroy(value)	
	}
}



type FfiConverterMapStringOptionalTypeMetadataValue struct {}

var FfiConverterMapStringOptionalTypeMetadataValueINSTANCE = FfiConverterMapStringOptionalTypeMetadataValue{}

func (c FfiConverterMapStringOptionalTypeMetadataValue) Lift(rb RustBufferI) map[string]*MetadataValue {
	return LiftFromRustBuffer[map[string]*MetadataValue](c, rb)
}

func (_ FfiConverterMapStringOptionalTypeMetadataValue) Read(reader io.Reader) map[string]*MetadataValue {
	result := make(map[string]*MetadataValue)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterOptionalTypeMetadataValueINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringOptionalTypeMetadataValue) Lower(value map[string]*MetadataValue) RustBuffer {
	return LowerIntoRustBuffer[map[string]*MetadataValue](c, value)
}

func (_ FfiConverterMapStringOptionalTypeMetadataValue) Write(writer io.Writer, mapValue map[string]*MetadataValue) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]*MetadataValue is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterOptionalTypeMetadataValueINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringOptionalTypeMetadataValue struct {}

func (_ FfiDestroyerMapStringOptionalTypeMetadataValue) Destroy(mapValue map[string]*MetadataValue) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerOptionalTypeMetadataValue{}.Destroy(value)	
	}
}



type FfiConverterMapStringSequenceTypeResourceIndicator struct {}

var FfiConverterMapStringSequenceTypeResourceIndicatorINSTANCE = FfiConverterMapStringSequenceTypeResourceIndicator{}

func (c FfiConverterMapStringSequenceTypeResourceIndicator) Lift(rb RustBufferI) map[string][]ResourceIndicator {
	return LiftFromRustBuffer[map[string][]ResourceIndicator](c, rb)
}

func (_ FfiConverterMapStringSequenceTypeResourceIndicator) Read(reader io.Reader) map[string][]ResourceIndicator {
	result := make(map[string][]ResourceIndicator)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterSequenceTypeResourceIndicatorINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringSequenceTypeResourceIndicator) Lower(value map[string][]ResourceIndicator) RustBuffer {
	return LowerIntoRustBuffer[map[string][]ResourceIndicator](c, value)
}

func (_ FfiConverterMapStringSequenceTypeResourceIndicator) Write(writer io.Writer, mapValue map[string][]ResourceIndicator) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string][]ResourceIndicator is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterSequenceTypeResourceIndicatorINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringSequenceTypeResourceIndicator struct {}

func (_ FfiDestroyerMapStringSequenceTypeResourceIndicator) Destroy(mapValue map[string][]ResourceIndicator) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerSequenceTypeResourceIndicator{}.Destroy(value)	
	}
}



type FfiConverterMapStringSequenceTypeResourceOrNonFungible struct {}

var FfiConverterMapStringSequenceTypeResourceOrNonFungibleINSTANCE = FfiConverterMapStringSequenceTypeResourceOrNonFungible{}

func (c FfiConverterMapStringSequenceTypeResourceOrNonFungible) Lift(rb RustBufferI) map[string][]ResourceOrNonFungible {
	return LiftFromRustBuffer[map[string][]ResourceOrNonFungible](c, rb)
}

func (_ FfiConverterMapStringSequenceTypeResourceOrNonFungible) Read(reader io.Reader) map[string][]ResourceOrNonFungible {
	result := make(map[string][]ResourceOrNonFungible)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterSequenceTypeResourceOrNonFungibleINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringSequenceTypeResourceOrNonFungible) Lower(value map[string][]ResourceOrNonFungible) RustBuffer {
	return LowerIntoRustBuffer[map[string][]ResourceOrNonFungible](c, value)
}

func (_ FfiConverterMapStringSequenceTypeResourceOrNonFungible) Write(writer io.Writer, mapValue map[string][]ResourceOrNonFungible) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string][]ResourceOrNonFungible is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterSequenceTypeResourceOrNonFungibleINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringSequenceTypeResourceOrNonFungible struct {}

func (_ FfiDestroyerMapStringSequenceTypeResourceOrNonFungible) Destroy(mapValue map[string][]ResourceOrNonFungible) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerSequenceTypeResourceOrNonFungible{}.Destroy(value)	
	}
}



type FfiConverterMapStringSequenceTypeResourceSpecifier struct {}

var FfiConverterMapStringSequenceTypeResourceSpecifierINSTANCE = FfiConverterMapStringSequenceTypeResourceSpecifier{}

func (c FfiConverterMapStringSequenceTypeResourceSpecifier) Lift(rb RustBufferI) map[string][]ResourceSpecifier {
	return LiftFromRustBuffer[map[string][]ResourceSpecifier](c, rb)
}

func (_ FfiConverterMapStringSequenceTypeResourceSpecifier) Read(reader io.Reader) map[string][]ResourceSpecifier {
	result := make(map[string][]ResourceSpecifier)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterSequenceTypeResourceSpecifierINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringSequenceTypeResourceSpecifier) Lower(value map[string][]ResourceSpecifier) RustBuffer {
	return LowerIntoRustBuffer[map[string][]ResourceSpecifier](c, value)
}

func (_ FfiConverterMapStringSequenceTypeResourceSpecifier) Write(writer io.Writer, mapValue map[string][]ResourceSpecifier) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string][]ResourceSpecifier is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterSequenceTypeResourceSpecifierINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringSequenceTypeResourceSpecifier struct {}

func (_ FfiDestroyerMapStringSequenceTypeResourceSpecifier) Destroy(mapValue map[string][]ResourceSpecifier) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerSequenceTypeResourceSpecifier{}.Destroy(value)	
	}
}



type FfiConverterMapStringMapStringTypeResourcePreferenceUpdate struct {}

var FfiConverterMapStringMapStringTypeResourcePreferenceUpdateINSTANCE = FfiConverterMapStringMapStringTypeResourcePreferenceUpdate{}

func (c FfiConverterMapStringMapStringTypeResourcePreferenceUpdate) Lift(rb RustBufferI) map[string]map[string]ResourcePreferenceUpdate {
	return LiftFromRustBuffer[map[string]map[string]ResourcePreferenceUpdate](c, rb)
}

func (_ FfiConverterMapStringMapStringTypeResourcePreferenceUpdate) Read(reader io.Reader) map[string]map[string]ResourcePreferenceUpdate {
	result := make(map[string]map[string]ResourcePreferenceUpdate)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterMapStringTypeResourcePreferenceUpdateINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringMapStringTypeResourcePreferenceUpdate) Lower(value map[string]map[string]ResourcePreferenceUpdate) RustBuffer {
	return LowerIntoRustBuffer[map[string]map[string]ResourcePreferenceUpdate](c, value)
}

func (_ FfiConverterMapStringMapStringTypeResourcePreferenceUpdate) Write(writer io.Writer, mapValue map[string]map[string]ResourcePreferenceUpdate) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]map[string]ResourcePreferenceUpdate is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterMapStringTypeResourcePreferenceUpdateINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringMapStringTypeResourcePreferenceUpdate struct {}

func (_ FfiDestroyerMapStringMapStringTypeResourcePreferenceUpdate) Destroy(mapValue map[string]map[string]ResourcePreferenceUpdate) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerMapStringTypeResourcePreferenceUpdate{}.Destroy(value)	
	}
}



type FfiConverterMapStringMapStringOptionalTypeMetadataValue struct {}

var FfiConverterMapStringMapStringOptionalTypeMetadataValueINSTANCE = FfiConverterMapStringMapStringOptionalTypeMetadataValue{}

func (c FfiConverterMapStringMapStringOptionalTypeMetadataValue) Lift(rb RustBufferI) map[string]map[string]*MetadataValue {
	return LiftFromRustBuffer[map[string]map[string]*MetadataValue](c, rb)
}

func (_ FfiConverterMapStringMapStringOptionalTypeMetadataValue) Read(reader io.Reader) map[string]map[string]*MetadataValue {
	result := make(map[string]map[string]*MetadataValue)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterMapStringOptionalTypeMetadataValueINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringMapStringOptionalTypeMetadataValue) Lower(value map[string]map[string]*MetadataValue) RustBuffer {
	return LowerIntoRustBuffer[map[string]map[string]*MetadataValue](c, value)
}

func (_ FfiConverterMapStringMapStringOptionalTypeMetadataValue) Write(writer io.Writer, mapValue map[string]map[string]*MetadataValue) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]map[string]*MetadataValue is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterMapStringOptionalTypeMetadataValueINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringMapStringOptionalTypeMetadataValue struct {}

func (_ FfiDestroyerMapStringMapStringOptionalTypeMetadataValue) Destroy(mapValue map[string]map[string]*MetadataValue) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerMapStringOptionalTypeMetadataValue{}.Destroy(value)	
	}
}



type FfiConverterMapTypePublicKeyFingerprintBytes struct {}

var FfiConverterMapTypePublicKeyFingerprintBytesINSTANCE = FfiConverterMapTypePublicKeyFingerprintBytes{}

func (c FfiConverterMapTypePublicKeyFingerprintBytes) Lift(rb RustBufferI) map[PublicKeyFingerprint][]byte {
	return LiftFromRustBuffer[map[PublicKeyFingerprint][]byte](c, rb)
}

func (_ FfiConverterMapTypePublicKeyFingerprintBytes) Read(reader io.Reader) map[PublicKeyFingerprint][]byte {
	result := make(map[PublicKeyFingerprint][]byte)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterTypePublicKeyFingerprintINSTANCE.Read(reader)
		value := FfiConverterBytesINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapTypePublicKeyFingerprintBytes) Lower(value map[PublicKeyFingerprint][]byte) RustBuffer {
	return LowerIntoRustBuffer[map[PublicKeyFingerprint][]byte](c, value)
}

func (_ FfiConverterMapTypePublicKeyFingerprintBytes) Write(writer io.Writer, mapValue map[PublicKeyFingerprint][]byte) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[PublicKeyFingerprint][]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterTypePublicKeyFingerprintINSTANCE.Write(writer, key)
		FfiConverterBytesINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapTypePublicKeyFingerprintBytes struct {}

func (_ FfiDestroyerMapTypePublicKeyFingerprintBytes) Destroy(mapValue map[PublicKeyFingerprint][]byte) {
	for key, value := range mapValue {
		FfiDestroyerTypePublicKeyFingerprint{}.Destroy(key)
		FfiDestroyerBytes{}.Destroy(value)	
	}
}



type FfiConverterMapTypeCurveTypeTypeDecryptorsByCurve struct {}

var FfiConverterMapTypeCurveTypeTypeDecryptorsByCurveINSTANCE = FfiConverterMapTypeCurveTypeTypeDecryptorsByCurve{}

func (c FfiConverterMapTypeCurveTypeTypeDecryptorsByCurve) Lift(rb RustBufferI) map[CurveType]DecryptorsByCurve {
	return LiftFromRustBuffer[map[CurveType]DecryptorsByCurve](c, rb)
}

func (_ FfiConverterMapTypeCurveTypeTypeDecryptorsByCurve) Read(reader io.Reader) map[CurveType]DecryptorsByCurve {
	result := make(map[CurveType]DecryptorsByCurve)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterTypeCurveTypeINSTANCE.Read(reader)
		value := FfiConverterTypeDecryptorsByCurveINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapTypeCurveTypeTypeDecryptorsByCurve) Lower(value map[CurveType]DecryptorsByCurve) RustBuffer {
	return LowerIntoRustBuffer[map[CurveType]DecryptorsByCurve](c, value)
}

func (_ FfiConverterMapTypeCurveTypeTypeDecryptorsByCurve) Write(writer io.Writer, mapValue map[CurveType]DecryptorsByCurve) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[CurveType]DecryptorsByCurve is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterTypeCurveTypeINSTANCE.Write(writer, key)
		FfiConverterTypeDecryptorsByCurveINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapTypeCurveTypeTypeDecryptorsByCurve struct {}

func (_ FfiDestroyerMapTypeCurveTypeTypeDecryptorsByCurve) Destroy(mapValue map[CurveType]DecryptorsByCurve) {
	for key, value := range mapValue {
		FfiDestroyerTypeCurveType{}.Destroy(key)
		FfiDestroyerTypeDecryptorsByCurve{}.Destroy(value)	
	}
}



type FfiConverterMapTypeEntityTypeSequenceAddress struct {}

var FfiConverterMapTypeEntityTypeSequenceAddressINSTANCE = FfiConverterMapTypeEntityTypeSequenceAddress{}

func (c FfiConverterMapTypeEntityTypeSequenceAddress) Lift(rb RustBufferI) map[EntityType][]*Address {
	return LiftFromRustBuffer[map[EntityType][]*Address](c, rb)
}

func (_ FfiConverterMapTypeEntityTypeSequenceAddress) Read(reader io.Reader) map[EntityType][]*Address {
	result := make(map[EntityType][]*Address)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterTypeEntityTypeINSTANCE.Read(reader)
		value := FfiConverterSequenceAddressINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapTypeEntityTypeSequenceAddress) Lower(value map[EntityType][]*Address) RustBuffer {
	return LowerIntoRustBuffer[map[EntityType][]*Address](c, value)
}

func (_ FfiConverterMapTypeEntityTypeSequenceAddress) Write(writer io.Writer, mapValue map[EntityType][]*Address) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[EntityType][]*Address is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterTypeEntityTypeINSTANCE.Write(writer, key)
		FfiConverterSequenceAddressINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapTypeEntityTypeSequenceAddress struct {}

func (_ FfiDestroyerMapTypeEntityTypeSequenceAddress) Destroy(mapValue map[EntityType][]*Address) {
	for key, value := range mapValue {
		FfiDestroyerTypeEntityType{}.Destroy(key)
		FfiDestroyerSequenceAddress{}.Destroy(value)	
	}
}




/**
 * Typealias from the type name used in the UDL file to the custom type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type HashableBytes = string

type FfiConverterTypeHashableBytes struct{}

var FfiConverterTypeHashableBytesINSTANCE = FfiConverterTypeHashableBytes{}

func (FfiConverterTypeHashableBytes) Lower(value HashableBytes) RustBufferI {
    builtinValue := []byte(value)
    return FfiConverterBytesINSTANCE.Lower(builtinValue)
}

func (FfiConverterTypeHashableBytes) Write(writer io.Writer, value HashableBytes) {
    builtinValue := []byte(value)
    FfiConverterBytesINSTANCE.Write(writer, builtinValue)
}

func (FfiConverterTypeHashableBytes) Lift(value RustBufferI) HashableBytes {
    builtinValue := FfiConverterBytesINSTANCE.Lift(value)
    return string(builtinValue)
}

func (FfiConverterTypeHashableBytes) Read(reader io.Reader) HashableBytes {
    builtinValue := FfiConverterBytesINSTANCE.Read(reader)
    return string(builtinValue)
}

type FfiDestroyerTypeHashableBytes struct {}

func (FfiDestroyerTypeHashableBytes) Destroy(value HashableBytes) {
	builtinValue := []byte(value)
	FfiDestroyerBytes{}.Destroy(builtinValue)
}

func DeriveOlympiaAccountAddressFromPublicKey(publicKey PublicKey, olympiaNetwork OlympiaNetwork) (*OlympiaAddress, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_olympia_account_address_from_public_key(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterTypeOlympiaNetworkINSTANCE.Lower(olympiaNetwork), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *OlympiaAddress
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterOlympiaAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func DerivePublicKeyFromOlympiaAccountAddress(olympiaResourceAddress *OlympiaAddress) (PublicKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_public_key_from_olympia_account_address(FfiConverterOlympiaAddressINSTANCE.Lower(olympiaResourceAddress), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue PublicKey
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypePublicKeyINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func DeriveResourceAddressFromOlympiaResourceAddress(olympiaResourceAddress *OlympiaAddress, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_resource_address_from_olympia_resource_address(FfiConverterOlympiaAddressINSTANCE.Lower(olympiaResourceAddress), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func DeriveVirtualAccountAddressFromOlympiaAccountAddress(olympiaAccountAddress *OlympiaAddress, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_virtual_account_address_from_olympia_account_address(FfiConverterOlympiaAddressINSTANCE.Lower(olympiaAccountAddress), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func DeriveVirtualAccountAddressFromPublicKey(publicKey PublicKey, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_virtual_account_address_from_public_key(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func DeriveVirtualIdentityAddressFromPublicKey(publicKey PublicKey, networkId uint8) (*Address, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_virtual_identity_address_from_public_key(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Address
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func DeriveVirtualSignatureNonFungibleGlobalIdFromPublicKey(publicKey PublicKey, networkId uint8) (*NonFungibleGlobalId, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_derive_virtual_signature_non_fungible_global_id_from_public_key(FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NonFungibleGlobalId
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNonFungibleGlobalIdINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func GetBuildInformation() BuildInformation {
	return FfiConverterTypeBuildInformationINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_get_build_information( _uniffiStatus)
	}))
}

func GetHash(data []byte) *Hash {
	return FfiConverterHashINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_get_hash(FfiConverterBytesINSTANCE.Lower(data), _uniffiStatus)
	}))
}

func GetKnownAddresses(networkId uint8) KnownAddresses {
	return FfiConverterTypeKnownAddressesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_get_known_addresses(FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	}))
}

func ManifestSborDecodeToStringRepresentation(bytes []byte, representation ManifestSborStringRepresentation, networkId uint8, schema *Schema) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_manifest_sbor_decode_to_string_representation(FfiConverterBytesINSTANCE.Lower(bytes), FfiConverterTypeManifestSborStringRepresentationINSTANCE.Lower(representation), FfiConverterUint8INSTANCE.Lower(networkId), FfiConverterOptionalTypeSchemaINSTANCE.Lower(schema), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func MetadataSborDecode(bytes []byte, networkId uint8) (MetadataValue, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_metadata_sbor_decode(FfiConverterBytesINSTANCE.Lower(bytes), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue MetadataValue
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeMetadataValueINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func MetadataSborEncode(value MetadataValue) ([]byte, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_metadata_sbor_encode(FfiConverterTypeMetadataValueINSTANCE.Lower(value), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func NonFungibleLocalIdAsStr(value NonFungibleLocalId) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_non_fungible_local_id_as_str(FfiConverterTypeNonFungibleLocalIdINSTANCE.Lower(value), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func NonFungibleLocalIdFromStr(string string) (NonFungibleLocalId, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_non_fungible_local_id_from_str(FfiConverterStringINSTANCE.Lower(string), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue NonFungibleLocalId
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeNonFungibleLocalIdINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func NonFungibleLocalIdSborDecode(bytes []byte) (NonFungibleLocalId, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_non_fungible_local_id_sbor_decode(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue NonFungibleLocalId
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeNonFungibleLocalIdINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func NonFungibleLocalIdSborEncode(value NonFungibleLocalId) ([]byte, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_non_fungible_local_id_sbor_encode(FfiConverterTypeNonFungibleLocalIdINSTANCE.Lower(value), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func PublicKeyFingerprintFromVec(bytes []byte) PublicKeyFingerprint {
	return FfiConverterTypePublicKeyFingerprintINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_public_key_fingerprint_from_vec(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	}))
}

func PublicKeyFingerprintToVec(value PublicKeyFingerprint) []byte {
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_public_key_fingerprint_to_vec(FfiConverterTypePublicKeyFingerprintINSTANCE.Lower(value), _uniffiStatus)
	}))
}

func SborDecodeToStringRepresentation(bytes []byte, representation SerializationMode, networkId uint8, schema *Schema) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_sbor_decode_to_string_representation(FfiConverterBytesINSTANCE.Lower(bytes), FfiConverterTypeSerializationModeINSTANCE.Lower(representation), FfiConverterUint8INSTANCE.Lower(networkId), FfiConverterOptionalTypeSchemaINSTANCE.Lower(schema), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func SborDecodeToTypedNativeEvent(eventTypeIdentifier EventTypeIdentifier, eventData []byte, networkId uint8) (TypedNativeEvent, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_sbor_decode_to_typed_native_event(FfiConverterTypeEventTypeIdentifierINSTANCE.Lower(eventTypeIdentifier), FfiConverterBytesINSTANCE.Lower(eventData), FfiConverterUint8INSTANCE.Lower(networkId), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue TypedNativeEvent
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeTypedNativeEventINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func ScryptoSborDecodeToStringRepresentation(bytes []byte, representation SerializationMode, networkId uint8, schema *Schema) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_scrypto_sbor_decode_to_string_representation(FfiConverterBytesINSTANCE.Lower(bytes), FfiConverterTypeSerializationModeINSTANCE.Lower(representation), FfiConverterUint8INSTANCE.Lower(networkId), FfiConverterOptionalTypeSchemaINSTANCE.Lower(schema), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func ScryptoSborEncodeStringRepresentation(representation ScryptoSborString) ([]byte, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_radix_engine_toolkit_uniffi_fn_func_scrypto_sbor_encode_string_representation(FfiConverterTypeScryptoSborStringINSTANCE.Lower(representation), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []byte
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

func TestPanic(message string) error {
	_, _uniffiErr := rustCallWithError(FfiConverterTypeRadixEngineToolkitError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_radix_engine_toolkit_uniffi_fn_func_test_panic(FfiConverterStringINSTANCE.Lower(message), _uniffiStatus)
		return false
	})
		return _uniffiErr
}

