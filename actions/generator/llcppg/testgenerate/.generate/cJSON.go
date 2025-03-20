package cjson

import (
	"github.com/goplus/llgo/c"
	"unsafe"
)

const CJSON_VERSION_MAJOR = 1
const CJSON_VERSION_MINOR = 7
const CJSON_VERSION_PATCH = 18
const CJSON_IsReference = 256
const CJSON_StringIsConst = 512
const CJSON_NESTING_LIMIT = 1000

/* The cJSON structure: */

type CJSON struct {
	Next        *CJSON
	Prev        *CJSON
	Child       *CJSON
	Type        c.Int
	Valuestring *int8
	Valueint    c.Int
	Valuedouble float64
	String      *int8
}

type CJSONHooks struct {
	MallocFn unsafe.Pointer
	FreeFn   unsafe.Pointer
}
type CJSONBool c.Int

/* returns the version of cJSON as a string */
//go:linkname CJSONVersion C.cJSON_Version
func CJSONVersion() *int8

/* Supply malloc, realloc and free functions to cJSON */
// llgo:link (*CJSONHooks).CJSONInitHooks C.cJSON_InitHooks
func (recv_ *CJSONHooks) CJSONInitHooks() {
}

/* Memory Management: the caller is always responsible to free the results from all variants of cJSON_Parse (with cJSON_Delete) and cJSON_Print (with stdlib free, cJSON_Hooks.free_fn, or cJSON_free as appropriate). The exception is cJSON_PrintPreallocated, where the caller has full responsibility of the buffer. */
/* Supply a block of JSON, and this returns a cJSON object you can interrogate. */
//go:linkname CJSONParse C.cJSON_Parse
func CJSONParse(value *int8) *CJSON

//go:linkname CJSONParseWithLength C.cJSON_ParseWithLength
func CJSONParseWithLength(value *int8, buffer_length uintptr) *CJSON

/* ParseWithOpts allows you to require (and check) that the JSON is null terminated, and to retrieve the pointer to the final byte parsed. */
/* If you supply a ptr in return_parse_end and parsing fails, then return_parse_end will contain a pointer to the error so will match cJSON_GetErrorPtr(). */
//go:linkname CJSONParseWithOpts C.cJSON_ParseWithOpts
func CJSONParseWithOpts(value *int8, return_parse_end **int8, require_null_terminated CJSONBool) *CJSON

//go:linkname CJSONParseWithLengthOpts C.cJSON_ParseWithLengthOpts
func CJSONParseWithLengthOpts(value *int8, buffer_length uintptr, return_parse_end **int8, require_null_terminated CJSONBool) *CJSON

/* Render a cJSON entity to text for transfer/storage. */
// llgo:link (*CJSON).CJSONPrint C.cJSON_Print
func (recv_ *CJSON) CJSONPrint() *int8 {
	return nil
}

/* Render a cJSON entity to text for transfer/storage without any formatting. */
// llgo:link (*CJSON).CJSONPrintUnformatted C.cJSON_PrintUnformatted
func (recv_ *CJSON) CJSONPrintUnformatted() *int8 {
	return nil
}

/* Render a cJSON entity to text using a buffered strategy. prebuffer is a guess at the final size. guessing well reduces reallocation. fmt=0 gives unformatted, =1 gives formatted */
// llgo:link (*CJSON).CJSONPrintBuffered C.cJSON_PrintBuffered
func (recv_ *CJSON) CJSONPrintBuffered(prebuffer c.Int, fmt CJSONBool) *int8 {
	return nil
}

/* Render a cJSON entity to text using a buffer already allocated in memory with given length. Returns 1 on success and 0 on failure. */
/* NOTE: cJSON is not always 100% accurate in estimating how much memory it will use, so to be safe allocate 5 bytes more than you actually need */
// llgo:link (*CJSON).CJSONPrintPreallocated C.cJSON_PrintPreallocated
func (recv_ *CJSON) CJSONPrintPreallocated(buffer *int8, length c.Int, format CJSONBool) CJSONBool {
	return 0
}

/* Delete a cJSON entity and all subentities. */
// llgo:link (*CJSON).CJSONDelete C.cJSON_Delete
func (recv_ *CJSON) CJSONDelete() {
}

/* Returns the number of items in an array (or object). */
// llgo:link (*CJSON).CJSONGetArraySize C.cJSON_GetArraySize
func (recv_ *CJSON) CJSONGetArraySize() c.Int {
	return 0
}

/* Retrieve item number "index" from array "array". Returns NULL if unsuccessful. */
// llgo:link (*CJSON).CJSONGetArrayItem C.cJSON_GetArrayItem
func (recv_ *CJSON) CJSONGetArrayItem(index c.Int) *CJSON {
	return nil
}

/* Get item "string" from object. Case insensitive. */
// llgo:link (*CJSON).CJSONGetObjectItem C.cJSON_GetObjectItem
func (recv_ *CJSON) CJSONGetObjectItem(string *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONGetObjectItemCaseSensitive C.cJSON_GetObjectItemCaseSensitive
func (recv_ *CJSON) CJSONGetObjectItemCaseSensitive(string *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONHasObjectItem C.cJSON_HasObjectItem
func (recv_ *CJSON) CJSONHasObjectItem(string *int8) CJSONBool {
	return 0
}

/* For analysing failed parses. This returns a pointer to the parse error. You'll probably need to look a few chars back to make sense of it. Defined when cJSON_Parse() returns 0. 0 when cJSON_Parse() succeeds. */
//go:linkname CJSONGetErrorPtr C.cJSON_GetErrorPtr
func CJSONGetErrorPtr() *int8

/* Check item type and return its value */
// llgo:link (*CJSON).CJSONGetStringValue C.cJSON_GetStringValue
func (recv_ *CJSON) CJSONGetStringValue() *int8 {
	return nil
}

// llgo:link (*CJSON).CJSONGetNumberValue C.cJSON_GetNumberValue
func (recv_ *CJSON) CJSONGetNumberValue() float64 {
	return 0
}

/* These functions check the type of an item */
// llgo:link (*CJSON).CJSONIsInvalid C.cJSON_IsInvalid
func (recv_ *CJSON) CJSONIsInvalid() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsFalse C.cJSON_IsFalse
func (recv_ *CJSON) CJSONIsFalse() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsTrue C.cJSON_IsTrue
func (recv_ *CJSON) CJSONIsTrue() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsBool C.cJSON_IsBool
func (recv_ *CJSON) CJSONIsBool() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsNull C.cJSON_IsNull
func (recv_ *CJSON) CJSONIsNull() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsNumber C.cJSON_IsNumber
func (recv_ *CJSON) CJSONIsNumber() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsString C.cJSON_IsString
func (recv_ *CJSON) CJSONIsString() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsArray C.cJSON_IsArray
func (recv_ *CJSON) CJSONIsArray() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsObject C.cJSON_IsObject
func (recv_ *CJSON) CJSONIsObject() CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONIsRaw C.cJSON_IsRaw
func (recv_ *CJSON) CJSONIsRaw() CJSONBool {
	return 0
}

/* These calls create a cJSON item of the appropriate type. */
//go:linkname CJSONCreateNull C.cJSON_CreateNull
func CJSONCreateNull() *CJSON

//go:linkname CJSONCreateTrue C.cJSON_CreateTrue
func CJSONCreateTrue() *CJSON

//go:linkname CJSONCreateFalse C.cJSON_CreateFalse
func CJSONCreateFalse() *CJSON

// llgo:link CJSONBool.CJSONCreateBool C.cJSON_CreateBool
func (recv_ CJSONBool) CJSONCreateBool() *CJSON {
	return nil
}

//go:linkname CJSONCreateNumber C.cJSON_CreateNumber
func CJSONCreateNumber(num float64) *CJSON

//go:linkname CJSONCreateString C.cJSON_CreateString
func CJSONCreateString(string *int8) *CJSON

/* raw json */
//go:linkname CJSONCreateRaw C.cJSON_CreateRaw
func CJSONCreateRaw(raw *int8) *CJSON

//go:linkname CJSONCreateArray C.cJSON_CreateArray
func CJSONCreateArray() *CJSON

//go:linkname CJSONCreateObject C.cJSON_CreateObject
func CJSONCreateObject() *CJSON

/* Create a string where valuestring references a string so
 * it will not be freed by cJSON_Delete */
//go:linkname CJSONCreateStringReference C.cJSON_CreateStringReference
func CJSONCreateStringReference(string *int8) *CJSON

/* Create an object/array that only references it's elements so
 * they will not be freed by cJSON_Delete */
// llgo:link (*CJSON).CJSONCreateObjectReference C.cJSON_CreateObjectReference
func (recv_ *CJSON) CJSONCreateObjectReference() *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONCreateArrayReference C.cJSON_CreateArrayReference
func (recv_ *CJSON) CJSONCreateArrayReference() *CJSON {
	return nil
}

/* These utilities create an Array of count items.
 * The parameter count cannot be greater than the number of elements in the number array, otherwise array access will be out of bounds.*/
//go:linkname CJSONCreateIntArray C.cJSON_CreateIntArray
func CJSONCreateIntArray(numbers *c.Int, count c.Int) *CJSON

//go:linkname CJSONCreateFloatArray C.cJSON_CreateFloatArray
func CJSONCreateFloatArray(numbers *float32, count c.Int) *CJSON

//go:linkname CJSONCreateDoubleArray C.cJSON_CreateDoubleArray
func CJSONCreateDoubleArray(numbers *float64, count c.Int) *CJSON

//go:linkname CJSONCreateStringArray C.cJSON_CreateStringArray
func CJSONCreateStringArray(strings **int8, count c.Int) *CJSON

/* Append item to the specified array/object. */
// llgo:link (*CJSON).CJSONAddItemToArray C.cJSON_AddItemToArray
func (recv_ *CJSON) CJSONAddItemToArray(item *CJSON) CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONAddItemToObject C.cJSON_AddItemToObject
func (recv_ *CJSON) CJSONAddItemToObject(string *int8, item *CJSON) CJSONBool {
	return 0
}

/* Use this when string is definitely const (i.e. a literal, or as good as), and will definitely survive the cJSON object.
 * WARNING: When this function was used, make sure to always check that (item->type & cJSON_StringIsConst) is zero before
 * writing to `item->string` */
// llgo:link (*CJSON).CJSONAddItemToObjectCS C.cJSON_AddItemToObjectCS
func (recv_ *CJSON) CJSONAddItemToObjectCS(string *int8, item *CJSON) CJSONBool {
	return 0
}

/* Append reference to item to the specified array/object. Use this when you want to add an existing cJSON to a new cJSON, but don't want to corrupt your existing cJSON. */
// llgo:link (*CJSON).CJSONAddItemReferenceToArray C.cJSON_AddItemReferenceToArray
func (recv_ *CJSON) CJSONAddItemReferenceToArray(item *CJSON) CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONAddItemReferenceToObject C.cJSON_AddItemReferenceToObject
func (recv_ *CJSON) CJSONAddItemReferenceToObject(string *int8, item *CJSON) CJSONBool {
	return 0
}

/* Remove/Detach items from Arrays/Objects. */
// llgo:link (*CJSON).CJSONDetachItemViaPointer C.cJSON_DetachItemViaPointer
func (recv_ *CJSON) CJSONDetachItemViaPointer(item *CJSON) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONDetachItemFromArray C.cJSON_DetachItemFromArray
func (recv_ *CJSON) CJSONDetachItemFromArray(which c.Int) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONDeleteItemFromArray C.cJSON_DeleteItemFromArray
func (recv_ *CJSON) CJSONDeleteItemFromArray(which c.Int) {
}

// llgo:link (*CJSON).CJSONDetachItemFromObject C.cJSON_DetachItemFromObject
func (recv_ *CJSON) CJSONDetachItemFromObject(string *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONDetachItemFromObjectCaseSensitive C.cJSON_DetachItemFromObjectCaseSensitive
func (recv_ *CJSON) CJSONDetachItemFromObjectCaseSensitive(string *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONDeleteItemFromObject C.cJSON_DeleteItemFromObject
func (recv_ *CJSON) CJSONDeleteItemFromObject(string *int8) {
}

// llgo:link (*CJSON).CJSONDeleteItemFromObjectCaseSensitive C.cJSON_DeleteItemFromObjectCaseSensitive
func (recv_ *CJSON) CJSONDeleteItemFromObjectCaseSensitive(string *int8) {
}

/* Update array items. */
// llgo:link (*CJSON).CJSONInsertItemInArray C.cJSON_InsertItemInArray
func (recv_ *CJSON) CJSONInsertItemInArray(which c.Int, newitem *CJSON) CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONReplaceItemViaPointer C.cJSON_ReplaceItemViaPointer
func (recv_ *CJSON) CJSONReplaceItemViaPointer(item *CJSON, replacement *CJSON) CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONReplaceItemInArray C.cJSON_ReplaceItemInArray
func (recv_ *CJSON) CJSONReplaceItemInArray(which c.Int, newitem *CJSON) CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONReplaceItemInObject C.cJSON_ReplaceItemInObject
func (recv_ *CJSON) CJSONReplaceItemInObject(string *int8, newitem *CJSON) CJSONBool {
	return 0
}

// llgo:link (*CJSON).CJSONReplaceItemInObjectCaseSensitive C.cJSON_ReplaceItemInObjectCaseSensitive
func (recv_ *CJSON) CJSONReplaceItemInObjectCaseSensitive(string *int8, newitem *CJSON) CJSONBool {
	return 0
}

/* Duplicate a cJSON item */
// llgo:link (*CJSON).CJSONDuplicate C.cJSON_Duplicate
func (recv_ *CJSON) CJSONDuplicate(recurse CJSONBool) *CJSON {
	return nil
}

/* Duplicate will create a new, identical cJSON item to the one you pass, in new memory that will
 * need to be released. With recurse!=0, it will duplicate any children connected to the item.
 * The item->next and ->prev pointers are always zero on return from Duplicate. */
/* Recursively compare two cJSON items for equality. If either a or b is NULL or invalid, they will be considered unequal.
 * case_sensitive determines if object keys are treated case sensitive (1) or case insensitive (0) */
// llgo:link (*CJSON).CJSONCompare C.cJSON_Compare
func (recv_ *CJSON) CJSONCompare(b *CJSON, case_sensitive CJSONBool) CJSONBool {
	return 0
}

/* Minify a strings, remove blank characters(such as ' ', '\t', '\r', '\n') from strings.
 * The input pointer json cannot point to a read-only address area, such as a string constant,
 * but should point to a readable and writable address area. */
//go:linkname CJSONMinify C.cJSON_Minify
func CJSONMinify(json *int8)

/* Helper functions for creating and adding items to an object at the same time.
 * They return the added item or NULL on failure. */
// llgo:link (*CJSON).CJSONAddNullToObject C.cJSON_AddNullToObject
func (recv_ *CJSON) CJSONAddNullToObject(name *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddTrueToObject C.cJSON_AddTrueToObject
func (recv_ *CJSON) CJSONAddTrueToObject(name *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddFalseToObject C.cJSON_AddFalseToObject
func (recv_ *CJSON) CJSONAddFalseToObject(name *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddBoolToObject C.cJSON_AddBoolToObject
func (recv_ *CJSON) CJSONAddBoolToObject(name *int8, boolean CJSONBool) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddNumberToObject C.cJSON_AddNumberToObject
func (recv_ *CJSON) CJSONAddNumberToObject(name *int8, number float64) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddStringToObject C.cJSON_AddStringToObject
func (recv_ *CJSON) CJSONAddStringToObject(name *int8, string *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddRawToObject C.cJSON_AddRawToObject
func (recv_ *CJSON) CJSONAddRawToObject(name *int8, raw *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddObjectToObject C.cJSON_AddObjectToObject
func (recv_ *CJSON) CJSONAddObjectToObject(name *int8) *CJSON {
	return nil
}

// llgo:link (*CJSON).CJSONAddArrayToObject C.cJSON_AddArrayToObject
func (recv_ *CJSON) CJSONAddArrayToObject(name *int8) *CJSON {
	return nil
}

/* helper for the cJSON_SetNumberValue macro */
// llgo:link (*CJSON).CJSONSetNumberHelper C.cJSON_SetNumberHelper
func (recv_ *CJSON) CJSONSetNumberHelper(number float64) float64 {
	return 0
}

/* Change the valuestring of a cJSON_String object, only takes effect when type of object is cJSON_String */
// llgo:link (*CJSON).CJSONSetValuestring C.cJSON_SetValuestring
func (recv_ *CJSON) CJSONSetValuestring(valuestring *int8) *int8 {
	return nil
}

/* malloc/free objects using the malloc/free functions that have been set with cJSON_InitHooks */
//go:linkname CJSONMalloc C.cJSON_malloc
func CJSONMalloc(size uintptr) unsafe.Pointer

//go:linkname CJSONFree C.cJSON_free
func CJSONFree(object unsafe.Pointer)
