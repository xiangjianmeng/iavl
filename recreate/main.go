package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/tendermint/iavl"
	db "github.com/tendermint/tm-db"
)

func main() {
	source, err := iavl.NewMutableTree(db.NewMemDB(), 0)
	if err != nil {
		log.Fatal(err)
	}

	// Generate random trees, using a fixed seed for reproducability
	rand.Seed(63720467)
	var version int64
	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			c := 97 + int64(math.Floor(rand.Float64()*26))
			key := []byte(fmt.Sprintf("%c", c))
			value := []byte{byte(c - 96)}
			source.Set(key, value)
		}
		_, version, err = source.SaveVersion()
		if err != nil {
			log.Fatal(err)
		}
	}
	dump(source)

	// Build a new tree and check hash
	target, err := iavl.NewMutableTree(db.NewMemDB(), 0)
	if err != nil {
		log.Fatal(err)
	}

	isource, err := source.GetImmutable(version)
	if err != nil {
		log.Fatal(err)
	}
	export, err := isource.Export()
	if err != nil {
		log.Fatal(err)
	}
	err = target.Import(version, export)
	if err != nil {
		log.Fatal(err)
	}
	dump(target)

	_, err = target.Export()
	if err != nil {
		log.Fatal(err)
	}
}

func dump(tree *iavl.MutableTree) {
	tree.Iterate(func(key []byte, value []byte) bool {
		fmt.Printf("%s:%s\n", string(key), hex.EncodeToString(value))
		return false
	})
	fmt.Printf("Hash: %s\n", hash(tree))
}

func hash(tree *iavl.MutableTree) string {
	versions := tree.AvailableVersions()
	version := int64(versions[len(versions)-1])
	itree, err := tree.GetImmutable(version)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(itree.Hash())
}
