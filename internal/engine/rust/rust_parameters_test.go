package rust

import (
	"strings"
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/engine"
	"github.com/ast-metrics/ast-metrics/pb"
)

// assertParameters checks the parameters of one function of the analyzed file,
// the names as much as the count: max_parameters_per_method reports the count,
// and a name counted twice is a parameter that does not exist.
func assertParameters(t *testing.T, src string, function string, expected ...string) {
	t.Helper()

	file, err := engine.CreateTestFileWithCode(&RustRunner{}, src)
	if err != nil {
		t.Fatalf("Could not analyze the source: %v", err)
	}

	fn := findFunction(file, function)
	if fn == nil {
		t.Fatalf("Function %s not found", function)
	}

	got := make([]string, 0, len(fn.Parameters))
	for _, p := range fn.Parameters {
		got = append(got, p.Name)
	}
	if strings.Join(got, ",") != strings.Join(expected, ",") {
		t.Errorf("Expected parameters [%s], got [%s]", strings.Join(expected, ", "), strings.Join(got, ", "))
	}
}

func findFunction(file *pb.File, name string) *pb.StmtFunction {
	candidates := append([]*pb.StmtFunction{}, file.Stmts.StmtFunction...)
	for _, class := range file.Stmts.StmtClass {
		candidates = append(candidates, class.Stmts.StmtFunction...)
	}
	for _, namespace := range file.Stmts.StmtNamespace {
		candidates = append(candidates, namespace.Stmts.StmtFunction...)
		for _, class := range namespace.Stmts.StmtClass {
			candidates = append(candidates, class.Stmts.StmtFunction...)
		}
	}
	for _, candidate := range candidates {
		if candidate.Name.Short == name {
			return candidate
		}
	}
	return nil
}

func Test_Types_Are_Not_Parameters(t *testing.T) {
	assertParameters(t, `
fn build(names: Vec<String>, label: &str, count: u8) {}
`, "build", "names", "label", "count")
}

func Test_Self_Counts_Once(t *testing.T) {
	assertParameters(t, `
struct Repository;
impl Repository {
    fn find(&self, identifier: u32, label: String) -> u32 { identifier }
}
`, "find", "self", "identifier", "label")
}
