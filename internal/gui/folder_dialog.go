package gui

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const (
	FOS_PICKFOLDERS     = 0x00000020
	FOS_FORCEFILESYSTEM = 0x00000040
	FOS_PATHMUSTEXIST   = 0x00000800
	SIGDN_FILESYSPATH   = 0x80058000
	HRESULT_CANCELLED   = 0x800704C7
)

var (
	CLSID_FILE_OPEN_DIALOG      = win.CLSID(mustGUID("{DC1C5A9C-E88A-4DDE-A5A1-60F82A20AEF7}"))
	IID_FILE_OPEN_DIALOG        = win.IID(mustGUID("{D57C7288-D4AD-4768-BE02-9D969532D960}"))
	IID_SHELL_ITEM              = win.IID(mustGUID("{43826D1E-E718-42EE-BC55-A1E261C37BFE}"))
	shell32                     = windows.NewLazySystemDLL("shell32.dll")
	shCreateItemFromParsingName = shell32.NewProc("SHCreateItemFromParsingName")
)

func mustGUID(value string) windows.GUID {
	guid, err := windows.GUIDFromString(value)
	if err != nil {
		panic(err)
	}

	return guid
}

type fileOpenDialog struct {
	vtable *fileOpenDialogVtbl
}

type fileOpenDialogVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	Show                uintptr
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
	GetResults          uintptr
	GetSelectedItems    uintptr
}

type shellItem struct {
	vtable *shellItemVtbl
}

type shellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
	GetAttributes  uintptr
	Compare        uintptr
}

func checkHRESULT(operation string, result uintptr) error {
	if int32(result) < 0 {
		return fmt.Errorf("%s failed (HRESULT 0x%08X)", operation, uint32(result))
	}
	return nil
}

func (dialog *fileOpenDialog) release() {
	syscall.SyscallN(dialog.vtable.Release, uintptr(unsafe.Pointer(dialog)))
}

func (item *shellItem) release() {
	syscall.SyscallN(item.vtable.Release, uintptr(unsafe.Pointer(item)))
}

func createShellItem(directory string) (*shellItem, error) {
	name, err := windows.UTF16PtrFromString(directory)
	if err != nil {
		return nil, err
	}

	var item *shellItem

	result, _, _ := shCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(name)),
		0,
		uintptr(unsafe.Pointer(&IID_SHELL_ITEM)),
		uintptr(unsafe.Pointer(&item)),
	)

	if err := checkHRESULT("SHCreateItemFromParsingName", result); err != nil {
		return nil, err
	}

	return item, nil
}

func browseFolder(owner *walk.MainWindow, title string, initialDirectory string) (string, bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hr := win.CoInitializeEx(nil, win.COINIT_APARTMENTTHREADED|win.COINIT_DISABLE_OLE1DDE)

	if err := checkHRESULT("CoInitializeEx", uintptr(hr)); err != nil {
		return "", false, err
	}
	defer win.CoUninitialize()

	var dialog *fileOpenDialog

	hr = win.CoCreateInstance(
		&CLSID_FILE_OPEN_DIALOG,
		nil,
		win.CLSCTX_INPROC_SERVER,
		&IID_FILE_OPEN_DIALOG,
		(*unsafe.Pointer)(unsafe.Pointer(&dialog)),
	)

	if err := checkHRESULT("CoCreateInstance", uintptr(hr)); err != nil {
		return "", false, err
	}
	defer dialog.release()

	var options uint32

	result, _, _ := syscall.SyscallN(
		dialog.vtable.GetOptions,
		uintptr(unsafe.Pointer(dialog)),
		uintptr(unsafe.Pointer(&options)),
	)

	if err := checkHRESULT("GetOptions", result); err != nil {
		return "", false, err
	}

	options |= FOS_PICKFOLDERS |
		FOS_FORCEFILESYSTEM |
		FOS_PATHMUSTEXIST

	result, _, _ = syscall.SyscallN(
		dialog.vtable.SetOptions,
		uintptr(unsafe.Pointer(dialog)),
		uintptr(options),
	)

	if err := checkHRESULT("SetOptions", result); err != nil {
		return "", false, err
	}

	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return "", false, err
	}

	result, _, _ = syscall.SyscallN(
		dialog.vtable.SetTitle,
		uintptr(unsafe.Pointer(dialog)),
		uintptr(unsafe.Pointer(titlePtr)),
	)

	if err := checkHRESULT("SetTitle", result); err != nil {
		return "", false, err
	}

	if initialDirectory != "" {
		if folder, err := createShellItem(initialDirectory); err == nil {
			syscall.SyscallN(
				dialog.vtable.SetFolder,
				uintptr(unsafe.Pointer(dialog)),
				uintptr(unsafe.Pointer(folder)),
			)
			folder.release()
		}
	}

	result, _, _ = syscall.SyscallN(
		dialog.vtable.Show,
		uintptr(unsafe.Pointer(dialog)),
		uintptr(owner.Handle()),
	)

	if uint32(result) == HRESULT_CANCELLED {
		return "", false, nil
	}

	if err := checkHRESULT("Show", result); err != nil {
		return "", false, err
	}

	var selectedItem *shellItem

	result, _, _ = syscall.SyscallN(
		dialog.vtable.GetResult,
		uintptr(unsafe.Pointer(dialog)),
		uintptr(unsafe.Pointer(&selectedItem)),
	)

	if err := checkHRESULT("GetResult", result); err != nil {
		return "", false, err
	}
	defer selectedItem.release()

	var selectedPath *uint16

	result, _, _ = syscall.SyscallN(
		selectedItem.vtable.GetDisplayName,
		uintptr(unsafe.Pointer(selectedItem)),
		SIGDN_FILESYSPATH,
		uintptr(unsafe.Pointer(&selectedPath)),
	)

	if err := checkHRESULT("GetDisplayName", result); err != nil {
		return "", false, err
	}
	defer win.CoTaskMemFree(
		uintptr(unsafe.Pointer(selectedPath)),
	)

	return windows.UTF16PtrToString(selectedPath), true, nil
}
