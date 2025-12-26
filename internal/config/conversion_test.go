package config

import (
	"reflect"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLuaToGo_Array(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test array conversion
	err := L.DoString(`
		return {1, 2, 3, 4, 5}
	`)
	if err != nil {
		t.Fatalf("Failed to execute Lua: %v", err)
	}

	result := luaToGo(L.Get(-1))
	L.Pop(1)

	// Should be []interface{}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	if len(arr) != 5 {
		t.Fatalf("Expected 5 elements, got %d", len(arr))
	}

	for i := 0; i < 5; i++ {
		expected := float64(i + 1)
		if arr[i] != expected {
			t.Errorf("arr[%d] = %v, want %v", i, arr[i], expected)
		}
	}
}

func TestLuaToGo_Map(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test map conversion
	err := L.DoString(`
		return {name = "voidabyss", version = "0.1.0"}
	`)
	if err != nil {
		t.Fatalf("Failed to execute Lua: %v", err)
	}

	result := luaToGo(L.Get(-1))
	L.Pop(1)

	// Should be map[string]interface{}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	if m["name"] != "voidabyss" {
		t.Errorf("name = %v, want 'voidabyss'", m["name"])
	}

	if m["version"] != "0.1.0" {
		t.Errorf("version = %v, want '0.1.0'", m["version"])
	}
}

func TestLuaToGo_MixedTable(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test mixed table (has both array elements and named keys) - should be map
	err := L.DoString(`
		return {1, 2, 3, name = "test"}
	`)
	if err != nil {
		t.Fatalf("Failed to execute Lua: %v", err)
	}

	result := luaToGo(L.Get(-1))
	L.Pop(1)

	// Should be map[string]interface{} because it has named keys
	_, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}
}

func TestLuaToGo_SparseArray(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test sparse array (missing elements) - should be map
	err := L.DoString(`
		local t = {}
		t[1] = "a"
		t[3] = "c"
		t[5] = "e"
		return t
	`)
	if err != nil {
		t.Fatalf("Failed to execute Lua: %v", err)
	}

	result := luaToGo(L.Get(-1))
	L.Pop(1)

	// Should be map because it's sparse
	_, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} for sparse array, got %T", result)
	}
}

func TestLuaToGo_NestedArray(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test nested arrays
	err := L.DoString(`
		return {{1, 2}, {3, 4}, {5, 6}}
	`)
	if err != nil {
		t.Fatalf("Failed to execute Lua: %v", err)
	}

	result := luaToGo(L.Get(-1))
	L.Pop(1)

	// Should be []interface{}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	if len(arr) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(arr))
	}

	// Check first nested array
	nested, ok := arr[0].([]interface{})
	if !ok {
		t.Fatalf("Expected nested []interface{}, got %T", arr[0])
	}

	if len(nested) != 2 {
		t.Fatalf("Expected 2 elements in nested array, got %d", len(nested))
	}

	if nested[0] != float64(1) || nested[1] != float64(2) {
		t.Errorf("nested[0] = %v, nested[1] = %v, want 1.0, 2.0", nested[0], nested[1])
	}
}

func TestGoToLua_Array(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test Go slice to Lua array
	goArray := []interface{}{1.0, 2.0, 3.0}
	luaVal := goToLua(L, goArray)

	tbl, ok := luaVal.(*lua.LTable)
	if !ok {
		t.Fatalf("Expected *lua.LTable, got %T", luaVal)
	}

	// Check MaxN (should be 3)
	if tbl.MaxN() != 3 {
		t.Errorf("MaxN() = %d, want 3", tbl.MaxN())
	}

	// Check values
	for i := 1; i <= 3; i++ {
		val := tbl.RawGetInt(i)
		num, ok := val.(lua.LNumber)
		if !ok {
			t.Errorf("tbl[%d] is not LNumber, got %T", i, val)
			continue
		}
		if float64(num) != float64(i) {
			t.Errorf("tbl[%d] = %f, want %f", i, float64(num), float64(i))
		}
	}
}

func TestGoToLua_Map(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test Go map to Lua table
	goMap := map[string]interface{}{
		"name":    "test",
		"version": 1.0,
	}
	luaVal := goToLua(L, goMap)

	tbl, ok := luaVal.(*lua.LTable)
	if !ok {
		t.Fatalf("Expected *lua.LTable, got %T", luaVal)
	}

	// Check name
	nameVal := L.GetField(tbl, "name")
	nameStr, ok := nameVal.(lua.LString)
	if !ok || string(nameStr) != "test" {
		t.Errorf("name = %v, want 'test'", nameVal)
	}

	// Check version
	versionVal := L.GetField(tbl, "version")
	versionNum, ok := versionVal.(lua.LNumber)
	if !ok || float64(versionNum) != 1.0 {
		t.Errorf("version = %v, want 1.0", versionVal)
	}
}

func TestLuaToGo_RoundTrip(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test round-trip conversion
	original := []interface{}{
		"string",
		123.0,
		true,
		[]interface{}{1.0, 2.0, 3.0},
		map[string]interface{}{"key": "value"},
	}

	// Go -> Lua
	luaVal := goToLua(L, original)

	// Lua -> Go
	result := luaToGo(luaVal)

	// Check result
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	if len(arr) != 5 {
		t.Fatalf("Expected 5 elements, got %d", len(arr))
	}

	// Check each element
	if arr[0] != "string" {
		t.Errorf("arr[0] = %v, want 'string'", arr[0])
	}

	if arr[1] != 123.0 {
		t.Errorf("arr[1] = %v, want 123.0", arr[1])
	}

	if arr[2] != true {
		t.Errorf("arr[2] = %v, want true", arr[2])
	}

	nested, ok := arr[3].([]interface{})
	if !ok {
		t.Errorf("arr[3] is not []interface{}, got %T", arr[3])
	} else if !reflect.DeepEqual(nested, []interface{}{1.0, 2.0, 3.0}) {
		t.Errorf("arr[3] = %v, want [1.0, 2.0, 3.0]", nested)
	}

	m, ok := arr[4].(map[string]interface{})
	if !ok {
		t.Errorf("arr[4] is not map[string]interface{}, got %T", arr[4])
	} else if m["key"] != "value" {
		t.Errorf("arr[4]['key'] = %v, want 'value'", m["key"])
	}
}
