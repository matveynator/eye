//go:build ( dragonfly || ios || freebsd || darwin || ( linux && ppc64 ) || ( linux && ppc64le )  || ( linux && s390x ) || ( linux && amd64 ) || ( linux && mips64 ) || ( linux && mips64le ) || ( linux && arm64 ) || ( linux && 386 ) || android || windows )

package Database

import (  _ "github.com/genjidb/genji/driver" )
