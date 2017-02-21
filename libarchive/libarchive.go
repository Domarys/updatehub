package libarchive

// FIXME: test this whole file

/*
#cgo pkg-config: libarchive
#include <archive.h>
#include <archive_entry.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"io"
	"unsafe"
)

type Archive struct {
	archive *C.struct_archive
}

type ArchiveEntry struct {
	entry *C.struct_archive_entry
}

type Api interface {
	NewRead() Archive
	ReadSupportFilterAll(a Archive)
	ReadSupportFormatRaw(a Archive)
	ReadSupportFormatEmpty(a Archive)
	ReadOpenFileName(a Archive, filename string, blockSize int) error
	ReadFree(a Archive)
	ReadNextHeader(a Archive, e ArchiveEntry) error
	ReadData(a Archive, buffer []byte, length int) (int, error)
}

type LibArchive struct {
}

func (la LibArchive) NewRead() Archive {
	a := Archive{}
	a.archive = C.archive_read_new()
	return a
}

func (la LibArchive) ReadSupportFilterAll(a Archive) {
	C.archive_read_support_filter_all(a.archive)
}

func (la LibArchive) ReadSupportFormatRaw(a Archive) {
	C.archive_read_support_format_raw(a.archive)
}

func (la LibArchive) ReadSupportFormatEmpty(a Archive) {
	C.archive_read_support_format_empty(a.archive)
}

func (la LibArchive) ReadOpenFileName(a Archive, filename string, blockSize int) error {
	cFilename := C.CString(filename)
	r := C.archive_read_open_filename(a.archive, cFilename, C.size_t(blockSize))
	C.free(unsafe.Pointer(cFilename))

	if r != C.ARCHIVE_OK {
		return fmt.Errorf(C.GoString(C.archive_error_string(a.archive)))
	}

	return nil
}

func (la LibArchive) ReadFree(a Archive) {
	C.archive_read_free(a.archive)
}

func (la LibArchive) ReadNextHeader(a Archive, e ArchiveEntry) error {
	r := C.archive_read_next_header(a.archive, &e.entry)

	if r == C.ARCHIVE_EOF {
		return io.EOF
	}

	if r != C.ARCHIVE_OK {
		return fmt.Errorf(C.GoString(C.archive_error_string(a.archive)))
	}

	return nil
}

func (la LibArchive) ReadData(a Archive, buffer []byte, length int) (int, error) {
	r := C.archive_read_data(a.archive, unsafe.Pointer(&buffer[0]), C.size_t(length))

	if r < 0 {
		return int(r), fmt.Errorf(C.GoString(C.archive_error_string(a.archive)))
	}

	return int(r), nil
}