package wire

import (
	"fmt"

	eventalepb "github.com/nohns/eventale/gen/v1"
)

func SemVerStr(pb *eventalepb.SemanticVersion) string {
	return fmt.Sprintf("v%d.%d.%d", pb.Major, pb.Minor, pb.Patch)
}
