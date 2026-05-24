package core

import (
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

var dictStore *data_structure.Dict

func init() {
	dictStore = data_structure.CreateDict()
}
