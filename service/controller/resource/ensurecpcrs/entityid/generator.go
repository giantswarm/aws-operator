package entityid

import (
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	// idChars represents the character set used to generate cluster IDs.
	// (does not contain 1 and l, to avoid confusion)
	idChars = "023456789abcdefghijkmnopqrstuvwxyz"

	// idLength represents the number of characters used to create a cluster ID.
	idLength = 5
)

var (
	// Use local instance of RNG. Can be overwritten with fixed seed in tests
	// if needed.
	localRng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func New() string {
	for {
		letterRunes := []rune(idChars)
		b := make([]rune, idLength)
		for i := range b {
			b[i] = letterRunes[localRng.Intn(len(letterRunes))]
		}

		id := string(b)

		if _, err := strconv.Atoi(id); err == nil {
			// string is numbers only, which we want to avoid
			continue
		}

		matched, err := regexp.MatchString("^[a-z]+$", id) // nolint:staticcheck
		if err != nil {
			panic(microerror.JSON(err))
		}

		if matched {
			// strings is letters only, which we also avoid
			continue
		}

		return id
	}
}
