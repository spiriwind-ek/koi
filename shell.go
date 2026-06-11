func printHelp() {
	fmt.Println(`Commands:
  help                  Show this help
  exit / quit           Exit the shell

Filesystem:
  fs.mkdir("/path")           Create directory
  fs.write("/path", val)      Write value
  fs.read("/path")            Read value
  fs.ls("/path")              List contents
  fs.rm("/path")              Delete
  fs.exists("/path")          Check existence

Matrix:
  math.mat_new("/data/A", 2, 2, {1,2,3,4})
  math.mat_mul("/data/A", "/data/B", "/data/C")
  math.mat_transpose("/data/A", "/data/AT")
  math.mat_determinant("/data/A")
  math.mat_inverse("/data/A", "/data/AI")
  math.mat_print("/data/A")
  math.mat_shape("/data/A")

Tensor:
  math.tensor_new("/data/T", {2,3,4}, {1,2,3,4,5,6,...})
  math.tensor_print("/data/T")
  math.tensor_shape("/data/T")

Vector:
  math.dot({1,2,3}, {4,5,6})     Dot product
  math.norm({1,2,3})             Norm
  math.cross({1,0,0}, {0,1,0})   Cross product (3D)

System:
  os.time()           Unix timestamp
  os.clock()          CPU time
  os.version()        Koi version
  os.edition()        Edition (full/lite)

Note: Paths must be quoted!
  ✗ fs.mkdir(test)
  ✓ fs.mkdir("/test")`)
}
