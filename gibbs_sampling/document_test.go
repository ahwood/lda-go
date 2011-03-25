package lda

import (
	"fmt"
	"testing"
)

const kNumTopics = 3
const kDocumentContent = "apple orange apple"
const kDocumentGoFmt = "&{[apple orange] [0 2] [0 0 0] [3 0 0]}"
const kCorpusFile = "testdata/corpus.txt"
const kCorpusGoFmt = "{[apple orange] [0 2] [0 0 0] [3 0]},{[jagar zebra] [0 1] [0 0] [2 0]}"

func TestNewDocument(t *testing.T) {
	if doc, _ := NewDocument("", kNumTopics); doc != nil {
		t.Errorf("NewDocument given empty text returns non-nil.")
	}
	if doc, _ := NewDocument("   ", kNumTopics); doc != nil {
		t.Errorf("NewDocument given whitespace-only text returns non-nil.")
	}
	if doc, _ := NewDocument("orange", kNumTopics); doc != nil {
		t.Errorf("NewDocument given a one-word text returns non-nil.")
	}
	if doc, err := NewDocument(kDocumentContent, kNumTopics); doc != nil {
		p := fmt.Sprintf("%v", doc)
		if p != kDocumentGoFmt {
			t.Errorf("Expecting %s, but got %s", kDocumentGoFmt, p)
		}
		if doc.Length() != 3 {
			t.Errorf("Expecting doc length = 3, but got %d", doc.Length())
		}
	} else {
		t.Errorf("Error parsing document: " + err.String())
	}
}

func TestWordIterator(t *testing.T) {
	doc, _ := NewDocument(kDocumentContent, kNumTopics)
	iter, _ := NewWordIterator(doc)
	if iter.Done() {
		t.Errorf("Unexpected iter.Done()")
	}
	if iter.Topic() != 0 || iter.Word() != "apple" {
		t.Errorf(fmt.Sprintf("iter.Topic() = %d, iter.Word() = %s.",
			iter.Topic(), iter.Word()));
	}
	if iter.Next(); iter.Done() {
		t.Errorf("Unexpected iter.Done()")
	}
	if iter.Topic() != 0 || iter.Word() != "apple" {
		t.Errorf(fmt.Sprintf("iter.Topic() = %d, iter.Word() = %s.",
			iter.Topic(), iter.Word()));
	}
	if iter.Next(); iter.Done() {
		t.Errorf("Unexpected iter.Done()")
	}
	if iter.Topic() != 0 || iter.Word() != "orange" {
		t.Errorf(fmt.Sprintf("iter.Topic() = %d, iter.Word() = %s.",
			iter.Topic(), iter.Word()));
	}
	if iter.Next(); !iter.Done() {
		t.Errorf("Expecting iter.Done(), but iter.Done() == false")
	}
}

func TestCorpus(t *testing.T) {
	corpus := NewCorpus()
	if len(*corpus) != 0 {
		t.Errorf("Empty corpus does not have length = 0")
	}
	doc, _ := NewDocument(kDocumentContent, kNumTopics)
	*corpus = append(*corpus, doc)
	if fmt.Sprintf("%v", (*corpus)[0]) != kDocumentGoFmt {
		t.Errorf("Expecting: " + kDocumentGoFmt +
			" but got: " + fmt.Sprintf("%v", (*corpus)[0]))
	}
}

func TestLoadCorpus(t *testing.T) {
	corpus, err := LoadCorpus(kCorpusFile, 2)
	if err != nil {
		t.Errorf("Error in loading: " + kCorpusFile + " : " + err.String())
	} else {
		corpus_gofmt := fmt.Sprintf("%v,%v", *(*corpus)[0], *(*corpus)[1])
		if corpus_gofmt != kCorpusGoFmt {
			t.Errorf("Expecting: " + kCorpusGoFmt + ", but got: " + corpus_gofmt)
		}
	}
}
