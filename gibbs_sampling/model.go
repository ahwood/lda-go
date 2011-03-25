package lda

import (
	"bufio"
	"fmt"
	"encoding/line"
	"os"
	"strings"
	"strconv"
)

const kMaxModelFileLineLength = 1024 * 1024 // at most 1MB per line

type Model struct {
	topic_histograms map[string]Histogram
	global_histogram Histogram
	zero_histogram   Histogram
}

// Create an empty model with num_topics topics.
func NewModel(num_topics int) *Model {
	model := new(Model)
	model.topic_histograms = make(map[string]Histogram)
	model.global_histogram = NewHistogram(num_topics)
	model.zero_histogram = NewHistogram(num_topics)
	return model
}

// Create a model by counting topic assignments in a corpus.
func CreateModel(num_topics int, corpus *Corpus) *Model {
	model := NewModel(num_topics)
	for _, v := range *corpus {
		for iter, _ := NewWordIterator(v); !iter.Done(); iter.Next() {
			model.IncrementTopic(iter.Word(), iter.Topic(), 1)
		}
	}
	return model
}

// Load model from a text file, which must be in the following format:
//
// word_0   N(word_0, topic_0)  N(word_0, topic_1) ...
// word_1   N(word_1, topic_0)  N(word_1, topic_1) ...
// ...
//
// where each line in the file is the topic histogram of a word,
// word_x is a string containing no whitespaces, and N(word_x,topic_y)
// is an integer, counting the number of times that word_x is assigned
// topic_y.  Fields in a line are separated by one or more whitespaces.
//
func LoadModel(filename string) (model *Model, err os.Error) {
	file, err := os.Open(filename, 0, 0)
	if err != nil {
		return nil, os.NewError("Cannot open file: " + filename)
	}
	defer file.Close()

	num_topics := 0
	model = new(Model)
	model.topic_histograms = make(map[string]Histogram)

	reader := line.NewReader(bufio.NewReader(file), kMaxModelFileLineLength)
	l, is_prefix, err := reader.ReadLine()
	for err == nil {
		line := string(l)

		if is_prefix {
			return nil, os.NewError("Encountered a long line:" + line)
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, os.NewError("Invalid line: " + line)
		}

		if _, present := model.topic_histograms[fields[0]]; present {
			return nil, os.NewError("Found duplicated word: " + fields[0])
		}

		if num_topics == 0 {
			num_topics = len(fields) - 1
			model.global_histogram = NewHistogram(num_topics)
			model.zero_histogram = NewHistogram(num_topics)
		} else if len(fields)-1 != num_topics {
			return nil, os.NewError("Inconsistent num_topics: " + line)
		}

		hist := NewHistogram(num_topics)
		var conv_err os.Error
		for i := 0; i < num_topics; i++ {
			hist[i], conv_err = strconv.Atoi(fields[i+1])
			if conv_err != nil {
				return nil, os.NewError("Failed conversion to int: " + fields[i+1])
			}
			model.global_histogram[i] += hist[i]
		}
		model.topic_histograms[fields[0]] = hist

		l, _, err = reader.ReadLine()
	}

	if err != os.EOF {
		return nil, os.NewError("Error reading: " + filename + err.String())
	}
	if len(model.topic_histograms) <= 0 {
		return nil, os.NewError("No valid line in file: " + filename)
	}

	return model, nil
}

func (model *Model) SaveModel(filename string) os.Error {
	file, err := os.Open(filename, os.O_WRONLY|os.O_CREAT|os.O_TRUNC, 0666)
	if err != nil {
		return os.NewError("Cannot open file: " + filename + " " + err.String())
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for k, v := range model.topic_histograms {
		fmt.Fprintf(writer, "%s", k)
		for _, c := range v {
			fmt.Fprintf(writer, " %d", c)
		}
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

func (model *Model) NumTopics() int {
	return len(model.global_histogram)
}

func (model *Model) NumWords() int {
	return len(model.topic_histograms)
}

func (model *Model) IncrementTopic(word string, topic int, count int) {
	if topic >= model.NumTopics() {
		panic(fmt.Sprintf("topic (%d) > num_topics (%d)",
			topic, model.NumTopics()))
	}
	if _, present := model.topic_histograms[word]; !present {
		model.topic_histograms[word] = NewHistogram(model.NumTopics())
	}

	model.topic_histograms[word][topic] += count
	model.global_histogram[topic] += count
}

func (model *Model) ReassignTopic(word string, old_topic int, new_topic int) {
	model.IncrementTopic(word, old_topic, -1)
	model.IncrementTopic(word, new_topic, 1)
}

func (model *Model) GetWordTopicHistogram(word string) Histogram {
	return model.topic_histograms[word];
}

func (model *Model) GetGlobalTopicHistogram() Histogram {
	return model.global_histogram;
}

func (model *Model) AccumulateModel(m *Model) {
	if model.NumTopics() != m.NumTopics() {
		panic(fmt.Sprintf("model has (%d) topics; m has (%d) topics.",
			model.NumTopics(), m.NumTopics()))
	}

	for word, v := range m.topic_histograms {
		if _, present := model.topic_histograms[word]; !present {
			model.topic_histograms[word] = NewHistogram(model.NumTopics())
		}
		for topic, c := range v {
			model.IncrementTopic(word, topic, c)
		}
	}
}
