package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/MichaelTJones/walk"
)

type fileInfo struct {
	path string
	name string
	size int64
	hash []byte
	next *fileInfo
}
type pairInfo struct {
	f1 *fileInfo
	f2 *fileInfo
}

func main() {
	root := "/home/egorse"
	size := int64(100000)
	noname := false

	//
	// Traversal filesystem
	//
	count := 0                             // count of files
	var files *fileInfo                    // list of files
	filesFn := func(ch chan interface{}) { // func to consume files from walkFn
		for f := range ch {
			f := f.(*fileInfo)
			count++
			f.next = files
			files = f
		}
	}
	ch1, done1 := Proc(filesFn)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// TODO Check if symlinks need some special treatment
		if info.IsDir() {
			name := info.Name()
			if name == ".git" { // skip by name
				return walk.SkipDir
			}

			return nil
		}

		s := info.Size()
		if size > s { // skip too small files
			return nil
		}

		rec := &fileInfo{
			path: path,
			name: info.Name(),
			size: s,
		}
		ch1 <- rec

		return nil
	}

	log.Printf("walking over %s", root)
	if err := walk.Walk(root, walkFn); err != nil {
		log.Fatal(err)
	}
	done1()
	log.Printf("detected %v files", count)

	//
	// compare files
	//
	same := make(map[string][]*fileInfo)
	sameFn := func(ch chan interface{}) {
		for p := range ch {
			p := p.(*pairInfo)
			key := fmt.Sprintf("%x", p.f1.hash)
			if same[key] == nil {
				same[key] = append(same[key], p.f1, p.f2)
			} else {
				same[key] = append(same[key], p.f2)
			}
		}
	}
	ch3, done3 := Proc(sameFn)

	pairsFn := func(ch chan interface{}) {
		for p := range ch {
			p := p.(*pairInfo)

			if p.f1.hash == nil {
				h, err := Hash(p.f1.path)
				if err != nil {
					log.Print(err)
				}
				p.f1.hash = h
			}
			if p.f2.hash == nil {
				h, err := Hash(p.f2.path)
				if err != nil {
					log.Print(err)
				}
				p.f2.hash = h
			}

			if p.f1.hash == nil || p.f2.hash == nil || !bytes.Equal(p.f1.hash, p.f2.hash) {
				continue
			}

			ch3 <- p
		}
	}
	ch2, done2 := Proc(pairsFn, 16)

	for f1 := files; f1 != nil; f1 = f1.next {
		for f2 := f1.next; f2 != nil; f2 = f2.next {
			if f1.size != f2.size {
				continue
			}
			if !noname && f1.name != f2.name {
				continue
			}

			if f1.path == f2.path {
				log.Fatalf("twice detected path %v", f1.path)
				panic(f1)
			}

			ch2 <- &pairInfo{f1, f2}
		}
	}
	done2()
	done3()
	log.Printf("detected %v groups", len(same))

	//
	// Generate output
	//
	groups := [][]*fileInfo{}
	for _, files := range same {
		sort.Slice(files, func(i, j int) bool {
			return len(files[i].path) < len(files[j].path)
		})
		groups = append(groups, files)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i][0].size > groups[j][0].size
	})
	for _, g := range groups {
		fmt.Printf("%v x%d %v\n", g[0].size, len(g), g[0].path)

		s := ""
		for _, f := range g[1:] {
			s = s + fmt.Sprintf("%s ", f.path)
		}
		fmt.Printf("!%s\n", s)
	}

	log.Printf("done!")
}
