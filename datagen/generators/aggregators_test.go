package generators_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/Joaquin6/mgodatagen/datagen/generators"
)

func TestNewAggregatorCond(t *testing.T) {

	newAggregatorTests := []struct {
		name    string
		config  generators.Config
		correct bool
	}{
		{
			name: "empty collection",
			config: generators.Config{
				Type:       generators.TypeCountAggregator,
				Query:      bson.M{"n": 1},
				Database:   "db",
				Collection: "",
			},
			correct: false,
		},
		{
			name: "empty field valueAggregator",
			config: generators.Config{
				Type:       generators.TypeValueAggregator,
				Collection: "coll",
				Query:      bson.M{"n": 1},
				Database:   "db",
				Field:      "",
			},
			correct: false,
		},
		{
			name: "empty field boundAggregator",
			config: generators.Config{
				Type:       generators.TypeBoundAggregator,
				Collection: "coll",
				Query:      bson.M{"n": 1},
				Database:   "db",
				Field:      "",
			},
			correct: false,
		},
		{
			name: "missing all",
			config: generators.Config{
				Type: generators.TypeCountAggregator,
			},
			correct: false,
		},
		{
			name: "empty query",
			config: generators.Config{
				Type:  generators.TypeCountAggregator,
				Query: bson.M{},
			},
			correct: false,
		},
		{
			name: "missing databse",
			config: generators.Config{
				Type:       generators.TypeBoundAggregator,
				Collection: "test",
				Query:      bson.M{"n": 1},
			},
			correct: false,
		},
	}

	ci := generators.NewCollInfo(1, []int{3, 4}, defaultSeed, nil, nil)

	for _, tt := range newAggregatorTests {
		t.Run(tt.name, func(t *testing.T) {
			var content = map[string]generators.Config{
				"k": tt.config,
			}
			_, err := ci.NewAggregatorSlice(content)
			if tt.correct && err != nil {
				t.Errorf("expected no error for config %v \nbut got \n%v", tt.config, err)
			} else if !tt.correct && err == nil {
				t.Errorf("expected an error for config %v but got none", tt.config)
			}
		})
	}
}

func TestNewAggregatorFromMap(t *testing.T) {

	contentList := loadCollConfig(t, "full-aggregation.json")

	documentAggregatorTests := []struct {
		name          string
		config        map[string]generators.Config
		correct       bool
		nBaggregators int
	}{
		{
			name: "empty collection name",
			config: map[string]generators.Config{
				"key": {
					Type:       generators.TypeValueAggregator,
					Collection: "",
				},
			},
			correct:       false,
			nBaggregators: 0,
		}, {
			name:          "full-aggregation.json[0]",
			config:        contentList[0],
			correct:       true,
			nBaggregators: 0,
		}, {
			name:          "full-aggregation.json[1]",
			config:        contentList[1],
			correct:       true,
			nBaggregators: 3,
		},
	}

	ci := generators.NewCollInfo(1, []int{3, 4}, defaultSeed, nil, nil)

	for _, tt := range documentAggregatorTests {
		t.Run(tt.name, func(t *testing.T) {
			aggs, err := ci.NewAggregatorSlice(tt.config)
			if tt.correct && err != nil {
				t.Errorf("expected no error for config %v \nbut got \n%v", tt.config, err)
			} else if !tt.correct && err == nil {
				t.Errorf("expected an error for config %v but got none", tt.config)
			}
			if want, got := tt.nBaggregators, len(aggs); want != got {
				t.Errorf("for config %v, expected %d agg but got %d", tt.config, want, got)
			}
		})
	}
}

func TestAggregatorUpdate(t *testing.T) {

	aggregatorUpdateTest := []struct {
		name           string
		baseDoc        []interface{}
		config         generators.Config
		expectedUpdate [2]bson.M
	}{
		{
			name: "countAggregator",
			baseDoc: []interface{}{
				bson.M{"_id": 1, "local": 1},
				bson.M{"_id": 2, "local": 2},
				bson.M{"_id": 3, "local": 1},
			},
			config: generators.Config{
				Type:       generators.TypeCountAggregator,
				Collection: "test",
				Database:   "datagen_it_test",
				Query: bson.M{
					"local": "$$_id",
				},
			},
			expectedUpdate: [2]bson.M{
				{"_id": 1},
				{"$set": bson.M{"key": int32(2)}},
			},
		},
		{
			name: "valueAggregator",
			baseDoc: []interface{}{
				bson.M{"_id": 1, "local": 1},
				bson.M{"_id": 2, "local": 1},
				bson.M{"_id": 3, "local": 2},
			},
			config: generators.Config{
				Type:       generators.TypeValueAggregator,
				Collection: "test",
				Database:   "datagen_it_test",
				Field:      "_id",
				Query: bson.M{
					"local": "$$_id",
				},
			},
			expectedUpdate: [2]bson.M{
				{"_id": 1},
				{"$set": bson.M{"key": []interface{}{1, 2}}},
			},
		},
		{
			name: "boundAggregator",
			baseDoc: []interface{}{
				bson.M{"_id": 1, "local": 2},
				bson.M{"_id": 2, "local": 1},
				bson.M{"_id": 3, "local": 1},
			},
			config: generators.Config{
				Type:       generators.TypeBoundAggregator,
				Collection: "test",
				Database:   "datagen_it_test",
				Field:      "_id",
				Query: bson.M{
					"local": "$$_id",
				},
			},
			expectedUpdate: [2]bson.M{
				{"_id": 1},
				{"$set": bson.M{"key": bson.M{"m": 2, "M": 3}}},
			},
		},
		{
			name: "countAggregator no local field",
			baseDoc: []interface{}{
				bson.M{"_id": 1, "field": 1},
				bson.M{"_id": 2, "field": 2},
			},
			config: generators.Config{
				Type:       generators.TypeCountAggregator,
				Collection: "test",
				Database:   "datagen_it_test",
				Query: bson.M{
					"field": 1,
				},
			},
			expectedUpdate: [2]bson.M{
				{"_id": 1},
				{"$set": bson.M{"key": int32(1)}},
			},
		},
	}

	ci := generators.NewCollInfo(1, []int{3, 4}, defaultSeed, nil, nil)
	session, err := mgo.Dial("mongodb://")
	if err != nil {
		t.Error(err)
	}

	for _, tt := range aggregatorUpdateTest {
		t.Run(tt.name, func(t *testing.T) {
			createCollection(t, session, tt.config, tt.baseDoc)
			aggregator := newAggregator(t, ci, tt.config)

			if want, got := aggregator.Query(), tt.config.Query; !reflect.DeepEqual(want, got) {
				t.Errorf("different keys, expected %s, got %s", want, got)
			}
			if want, got := aggregator.LocalVar(), "_id"; want != got {
				t.Errorf("different keys, expected %s, got %s", want, got)
			}

			update, err := aggregator.Update(session, 1)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(tt.expectedUpdate, update) {
				t.Errorf("expected %v, got %v", tt.expectedUpdate, update)
			}
		})
	}
}

func newAggregator(t *testing.T, ci *generators.CollInfo, config generators.Config) generators.Aggregator {
	var content = map[string]generators.Config{
		"key": config,
	}
	aggregators, err := ci.NewAggregatorSlice(content)
	if err != nil {
		t.Error(err)
	}
	if len(aggregators) == 0 {
		t.Error(fmt.Errorf("no aggregator created"))
	}
	return aggregators[0]
}

func createCollection(t *testing.T, session *mgo.Session, config generators.Config, baseDoc []interface{}) {
	coll := session.DB(config.Database).C(config.Collection)
	coll.RemoveAll(nil)
	err := coll.Insert(baseDoc...)
	if err != nil {
		t.Error(err)
	}
}
