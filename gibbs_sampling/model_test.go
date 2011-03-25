package lda

import (
	"fmt"
	"testing"
)

const kTestModelFile = "testdata/model.txt"
const kTestModelEncoding = "{map[orange:[0 1] zebra:[1 0] monky:[1 0] apple:[0 1] banana:[0 1]] [2 3] [0 0]}"
const kTmpModelFile = "/tmp/tmp_model.txt"

func TestLoadModel(t *testing.T) {
	model, err := LoadModel(kTestModelFile)
	if err != nil {
		t.Errorf("Unexpected error in loading: " + kTestModelFile + " due to " + err.String())
	} else {
		encoding := fmt.Sprintf("%v", *model)
		if encoding != kTestModelEncoding {
			t.Errorf(fmt.Sprintf("Expecting: %s\nbut got: %s",
				kTestModelEncoding, encoding))
		}
	}
}

func TestReassignTopic(t *testing.T) {
	model := NewModel(2)
	model.IncrementTopic("apple", 0, 3)
	model.IncrementTopic("orange", 1, 5)
	model.ReassignTopic("apple", 0, 1)
	if fmt.Sprintf("%v", *model) != "{map[orange:[0 5] apple:[2 1]] [2 6] [0 0]}" {
		t.Errorf("Unexpected model: %v", *model)
	}
}

func TestSaveModel(t *testing.T) {
	model, err := LoadModel(kTestModelFile)
	if err != nil {
		t.Errorf("Unexpected error in loading: " + kTestModelFile + " due to " + err.String())
	} else {
		err = model.SaveModel(kTmpModelFile)
		if err != nil {
			t.Errorf("Cannot write to: " + kTmpModelFile + " due to " + err.String())
		}

		model_new, err := LoadModel(kTmpModelFile)
		if err != nil {
			t.Errorf("Unexpected error in loading: " + kTmpModelFile + " due to " + err.String())
		} else {
			if fmt.Sprintf("%v", *model) != fmt.Sprintf("%v", *model_new) {
				t.Errorf("Original model: %v\ndoes not equal to loaded&saved model:%v",
					*model, *model_new)
			}
		}
	}
}

func TestCreateModel(t *testing.T) {
	corpus := NewCorpus()
	doc, _ := NewDocument("apple orange apple", 2)
	*corpus = append(*corpus, doc)
	doc, _ = NewDocument("zebra cat", 2)
	*corpus = append(*corpus, doc)
	model := CreateModel(2, corpus)

	const kModelGoFmt = "{map[orange:[1 0] zebra:[1 0] apple:[2 0] cat:[1 0]] [5 0] [0 0]}"
	if fmt.Sprintf("%v", *model) != kModelGoFmt {
		t.Errorf("Expecting: " + kModelGoFmt + ", but got: " + fmt.Sprintf("%v", *model))
	}
}

func TestAccumulateModel(t *testing.T) {
	model_1 := NewModel(2)
	model_1.IncrementTopic("apple", 0, 1);
	model_1.IncrementTopic("orange", 0, 1);

	model_2 := NewModel(2)
	model_2.IncrementTopic("orange", 1, 1);
	model_2.AccumulateModel(model_1)

	const kModelGoFmt = "{map[orange:[1 1] apple:[1 0]] [2 1] [0 0]}"
	if fmt.Sprintf("%v", *model_2) != kModelGoFmt {
		t.Errorf("Expecting: " + kModelGoFmt + ", but got: " + fmt.Sprintf("%v", *model_2))
	}
}
