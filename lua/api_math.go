package lua

import (
	"encoding/json"
	"fmt"
	"log"

	lua "github.com/yuin/gopher-lua"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	"koi/config"
	"koi/storage"
)

// math.mat_new(path, rows, cols, data_table)
func mathMatNew(db *storage.DB, cfg *config.Config) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		rows := L.CheckInt(2)
		cols := L.CheckInt(3)
		dataTbl := L.CheckTable(4)

		// Security: check matrix size limit
		maxSize := cfg.Security.MaxMatrixSize
		if maxSize > 0 && (rows > maxSize || cols > maxSize) {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("matrix size %dx%d exceeds limit %d", rows, cols, maxSize)))
			return 2
		}

		data := make([]float64, rows*cols)
		for i := 0; i < rows*cols; i++ {
			val := dataTbl.RawGetInt(i + 1)
			if num, ok := val.(lua.LNumber); ok {
				data[i] = float64(num)
			}
		}

		shape := fmt.Sprintf("[%d,%d]", rows, cols)
		dataJSON, err := json.Marshal(data)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("marshal data: %v", err)))
			return 2
		}
		meta := fmt.Sprintf(`{"format":"dense","shape":%s}`, shape)

		parent := parentPath(path)
		if !db.NodeExists(parent) {
			db.AutoMkdir(parent)
		}

		n := &storage.Node{
			Key:      path,
			ParentKey: &parent,
			Name:     baseName(path),
			ObjType:  "matrix",
			Value:    strPtr(string(dataJSON)),
			BlobMeta: &meta,
		}

		if db.NodeExists(path) {
			if err := db.UpdateNode(n); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		} else {
			if err := db.CreateNode(n); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// math.mat_mul(path_a, path_b, path_result)
func mathMatMul(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		aPath := L.CheckString(1)
		bPath := L.CheckString(2)
		resultPath := L.CheckString(3)

		A, err := loadMatrix(db, aPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		B, err := loadMatrix(db, bPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		ar, ac := A.Dims()
		br, bc := B.Dims()
		if ac != br {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("dimension mismatch: %dx%d * %dx%d", ar, ac, br, bc)))
			return 2
		}

		C := mat.NewDense(ar, bc, nil)
		C.Mul(A, B)

		saveMatrix(db, resultPath, C)

		L.Push(lua.LTrue)
		return 1
	}
}

// math.mat_transpose(path, result_path)
func mathMatTranspose(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		resultPath := L.CheckString(2)

		A, err := loadMatrix(db, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		r, c := A.Dims()
		T := mat.NewDense(c, r, nil)
		T.Copy(A.T())

		saveMatrix(db, resultPath, T)

		L.Push(lua.LTrue)
		return 1
	}
}

// math.mat_determinant(path) -> number
func mathMatDeterminant(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		A, err := loadMatrix(db, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		r, c := A.Dims()
		if r != c {
			L.Push(lua.LNil)
			L.Push(lua.LString("determinant requires square matrix"))
			return 2
		}

		det := mat.Det(A)
		L.Push(lua.LNumber(det))
		return 1
	}
}

// math.mat_inverse(path, result_path)
func mathMatInverse(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		resultPath := L.CheckString(2)

		A, err := loadMatrix(db, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		r, c := A.Dims()
		if r != c {
			L.Push(lua.LNil)
			L.Push(lua.LString("inverse requires square matrix"))
			return 2
		}

		inv := mat.NewDense(r, c, nil)
		if err := inv.Inverse(A); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		saveMatrix(db, resultPath, inv)

		L.Push(lua.LTrue)
		return 1
	}
}

// math.mat_print(path) -> string
func mathMatPrint(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		A, err := loadMatrix(db, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(fmt.Sprintf("%v", mat.Formatted(A, mat.Prefix(""), mat.Squeeze()))))
		return 1
	}
}

// math.mat_shape(path) -> table {rows, cols}
func mathMatShape(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		A, err := loadMatrix(db, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		r, c := A.Dims()
		tbl := L.NewTable()
		tbl.Append(lua.LNumber(r))
		tbl.Append(lua.LNumber(c))

		L.Push(tbl)
		return 1
	}
}

// math.tensor_new(path, shape_table, data_table)
func mathTensorNew(db *storage.DB, cfg *config.Config) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		shapeTbl := L.CheckTable(2)
		dataTbl := L.CheckTable(3)

		shape := tableToInts(L, shapeTbl)

		// Security: check tensor dimension limit
		maxNdim := cfg.Security.MaxTensorNdim
		if maxNdim > 0 && len(shape) > maxNdim {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("tensor ndim %d exceeds limit %d", len(shape), maxNdim)))
			return 2
		}

		totalSize := 1
		for _, s := range shape {
			totalSize *= s
		}

		data := make([]float64, totalSize)
		for i := 0; i < totalSize; i++ {
			val := dataTbl.RawGetInt(i + 1)
			if num, ok := val.(lua.LNumber); ok {
				data[i] = float64(num)
			}
		}

		dataJSON, err := json.Marshal(data)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("marshal data: %v", err)))
			return 2
		}
		shapeJSON, _ := json.Marshal(shape)
		meta := fmt.Sprintf(`{"format":"tensor","shape":%s,"ndim":%d}`, string(shapeJSON), len(shape))

		parent := parentPath(path)
		if !db.NodeExists(parent) {
			db.AutoMkdir(parent)
		}

		n := &storage.Node{
			Key:      path,
			ParentKey: &parent,
			Name:     baseName(path),
			ObjType:  "tensor",
			Value:    strPtr(string(dataJSON)),
			BlobMeta: &meta,
		}

		if db.NodeExists(path) {
			if err := db.UpdateNode(n); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		} else {
			if err := db.CreateNode(n); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// math.tensor_print(path) -> string
func mathTensorPrint(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		n, err := db.GetNode(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		var data []float64
		if err := json.Unmarshal([]byte(*n.Value), &data); err != nil {
			log.Printf("tensor_print unmarshal error: %v", err)
		}

		var shape []int
		if n.BlobMeta != nil {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(*n.BlobMeta), &meta); err != nil {
				log.Printf("tensor_print meta unmarshal error: %v", err)
			} else if s, ok := meta["shape"].([]interface{}); ok {
				for _, v := range s {
					if f, ok := v.(float64); ok {
						shape = append(shape, int(f))
					}
				}
			}
		}

		L.Push(lua.LString(fmt.Sprintf("tensor shape=%v data=%v", shape, data)))
		return 1
	}
}

// math.tensor_shape(path) -> table
func mathTensorShape(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		n, err := db.GetNode(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		var shape []int
		if n.BlobMeta != nil {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(*n.BlobMeta), &meta); err != nil {
				log.Printf("tensor_shape meta unmarshal error: %v", err)
			} else if s, ok := meta["shape"].([]interface{}); ok {
				for _, v := range s {
					if f, ok := v.(float64); ok {
						shape = append(shape, int(f))
					}
				}
			}
		}

		tbl := L.NewTable()
		for _, s := range shape {
			tbl.Append(lua.LNumber(s))
		}

		L.Push(tbl)
		return 1
	}
}

// math.dot(a_table, b_table) -> number
func mathDot(L *lua.LState) int {
	aTbl := L.CheckTable(1)
	bTbl := L.CheckTable(2)

	a := tableToFloats(L, aTbl)
	b := tableToFloats(L, bTbl)

	if len(a) != len(b) {
		L.Push(lua.LNil)
		L.Push(lua.LString("dimension mismatch"))
		return 2
	}

	L.Push(lua.LNumber(floats.Dot(a, b)))
	return 1
}

// math.norm(a_table) -> number
func mathNorm(L *lua.LState) int {
	tbl := L.CheckTable(1)
	a := tableToFloats(L, tbl)
	L.Push(lua.LNumber(floats.Norm(a, 2)))
	return 1
}

// math.cross(a_table, b_table) -> table (3D cross product)
func mathCross(L *lua.LState) int {
	aTbl := L.CheckTable(1)
	bTbl := L.CheckTable(2)

	a := tableToFloats(L, aTbl)
	b := tableToFloats(L, bTbl)

	if len(a) != 3 || len(b) != 3 {
		L.Push(lua.LNil)
		L.Push(lua.LString("cross product requires 3D vectors"))
		return 2
	}

	result := []float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}

	tbl := L.NewTable()
	for _, v := range result {
		tbl.Append(lua.LNumber(v))
	}

	L.Push(tbl)
	return 1
}

// helpers
func loadMatrix(db *storage.DB, path string) (*mat.Dense, error) {
	n, err := db.GetNode(path)
	if err != nil {
		return nil, fmt.Errorf("node not found: %s", path)
	}
	if n.ObjType != "matrix" {
		return nil, fmt.Errorf("not a matrix: %s", path)
	}

	var data []float64
	if err := json.Unmarshal([]byte(*n.Value), &data); err != nil {
		return nil, fmt.Errorf("parse data: %w", err)
	}

	var shape []int
	if n.BlobMeta != nil {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(*n.BlobMeta), &meta); err != nil {
			return nil, fmt.Errorf("parse meta: %w", err)
		}
		if s, ok := meta["shape"].([]interface{}); ok {
			for _, v := range s {
				if f, ok := v.(float64); ok {
					shape = append(shape, int(f))
				}
			}
		}
	}

	if len(shape) != 2 {
		return nil, fmt.Errorf("invalid shape for matrix")
	}

	if len(data) != shape[0]*shape[1] {
		return nil, fmt.Errorf("data length %d != shape %dx%d", len(data), shape[0], shape[1])
	}

	return mat.NewDense(shape[0], shape[1], data), nil
}

func saveMatrix(db *storage.DB, path string, A *mat.Dense) {
	r, c := A.Dims()
	data := A.RawMatrix().Data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Printf("saveMatrix marshal error: %v", err)
		return
	}
	shape := fmt.Sprintf("[%d,%d]", r, c)
	meta := fmt.Sprintf(`{"format":"dense","shape":%s}`, shape)

	parent := parentPath(path)
	if !db.NodeExists(parent) {
		db.AutoMkdir(parent)
	}

	n := &storage.Node{
		Key:      path,
		ParentKey: &parent,
		Name:     baseName(path),
		ObjType:  "matrix",
		Value:    strPtr(string(dataJSON)),
		BlobMeta: &meta,
	}

	if db.NodeExists(path) {
		if err := db.UpdateNode(n); err != nil {
			log.Printf("saveMatrix update error: %v", err)
		}
	} else {
		if err := db.CreateNode(n); err != nil {
			log.Printf("saveMatrix create error: %v", err)
		}
	}
}

func strPtr(s string) *string {
	return &s
}

func tableToFloats(L *lua.LState, tbl *lua.LTable) []float64 {
	var result []float64
	tbl.ForEach(func(_, value lua.LValue) {
		if num, ok := value.(lua.LNumber); ok {
			result = append(result, float64(num))
		}
	})
	return result
}

func tableToInts(L *lua.LState, tbl *lua.LTable) []int {
	var result []int
	tbl.ForEach(func(_, value lua.LValue) {
		if num, ok := value.(lua.LNumber); ok {
			result = append(result, int(num))
		}
	})
	return result
}
