package generators_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/Joaquin6/mgodatagen/datagen/generators"
)

const defaultSeed = 0

func TestIsDocumentCorrect(t *testing.T) {

	var expectedDoc struct {
		ID                      bson.ObjectId `bson:"_id"`
		UUID                    string        `bson:"uuid"`
		String                  string        `bson:"string"`
		Int32                   int32         `bson:"int32"`
		Int64                   int64         `bson:"int64"`
		Float                   float64       `bson:"float"`
		ConstInt32              int32         `bson:"constInt32"`
		ConstInt64              int64         `bson:"constInt64"`
		ConstFloat              float64       `bson:"constFloat"`
		Boolean                 bool          `bson:"boolean"`
		Position                []float64     `bson:"position"`
		StringFromArray         string        `bson:"stringFromArray"`
		IntFromArrayRandomOrder int           `bson:"intFromArrayRandomOrder"`
		ArrayFromArray          []string      `bson:"arrayFromArray"`
		ConstArray              []string      `bson:"constArray"`
		Fake                    string        `bson:"faker"`
		Constant                int32         `bson:"constant"`
		AutoIncrementInt32      int32         `bson:"autoIncrementInt32"`
		AutoIncrementInt64      int64         `bson:"autoIncrementInt64"`
		Date                    time.Time     `bson:"date"`
		BinaryData              []byte        `bson:"binaryData"`
		ArrayInt32              []int32       `bson:"arrayInt32"`
		Object                  struct {
			K1    string `bson:"k1"`
			K2    int32  `bson:"k2"`
			Subob struct {
				Sk int32 `bson:"s-k"`
			} `bson:"sub-ob"`
		} `bson:"object"`
	}

	fullDocumentTests := []struct {
		name     string
		content  map[string]generators.Config
		expected interface{}
	}{
		{
			name:     "full-bson.json",
			content:  loadCollConfig(t, "full-bson.json")[0],
			expected: expectedDoc,
		},
		{
			name:     "empty generator",
			content:  map[string]generators.Config{},
			expected: bson.M{},
		},
	}

	ci := generators.NewCollInfo(1000, []int{3, 2}, defaultSeed, map[int][][]byte{}, map[int]byte{})

	for _, tt := range fullDocumentTests {
		t.Run(tt.name, func(t *testing.T) {
			docGenerator, err := ci.NewDocumentGenerator(tt.content)
			if err != nil {
				t.Error(err)
			}
			for i := 0; i < ci.Count; i++ {
				err := bson.Unmarshal(docGenerator.Generate(), &tt.expected)
				if err != nil {
					t.Errorf("fail to unmarshal doc: %v", err)
				}
			}
		})
	}
}

func TestBigArray(t *testing.T) {

	ci := generators.NewCollInfo(-1, []int{3, 6, 4}, defaultSeed, nil, nil)
	docGenerator, err := ci.NewDocumentGenerator(map[string]generators.Config{
		"key": {Type: generators.TypeArray, Size: 15, ArrayContent: &generators.Config{Type: generators.TypeBoolean}},
	})
	if err != nil {
		t.Error(err)
	}

	var a struct {
		Key []bool `bson:"key"`
	}
	for i := 0; i < 100; i++ {
		err := bson.Unmarshal(docGenerator.Generate(), &a)
		if err != nil {
			t.Error(err)
		}
		if want, got := 15, len(a.Key); want != got {
			t.Errorf("wrong array size, expected %d, got %d", want, got)
		}
	}
}

func TestDocumentWithDecimal128(t *testing.T) {

	ci := generators.NewCollInfo(1, []int{3, 6, 4}, defaultSeed, nil, nil)
	docGenerator, err := ci.NewDocumentGenerator(map[string]generators.Config{
		"key": {Type: generators.TypeDecimal},
	})
	if err != nil {
		t.Error(err)
	}

	var d struct {
		Decimal bson.Decimal128 `bson:"decimal"`
	}
	for i := 0; i < 1000; i++ {
		err := bson.Unmarshal(docGenerator.Generate(), &d)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestNewGenerator(t *testing.T) {
	type testCase struct {
		name    string
		config  generators.Config
		correct bool
		version []int
	}

	newGeneratorTests := []testCase{
		{
			name: "string invalid minLength",
			config: generators.Config{
				Type:      generators.TypeString,
				MinLength: -1,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "string invalid maxLength",
			config: generators.Config{
				Type:      generators.TypeString,
				MinLength: 5,
				MaxLength: 2,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "unique with length == 0",
			config: generators.Config{
				Type:      generators.TypeString,
				MinLength: 0,
				Unique:    true,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "unique with string size to low",
			config: generators.Config{
				Type:      generators.TypeString,
				MinLength: 1,
				MaxLength: 1,
				Unique:    true,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "maxDistinctValue too high",
			config: generators.Config{
				Type:             generators.TypeString,
				MinLength:        0,
				MaxDistinctValue: 10,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "int with missing MinInt and MaxInt",
			config: generators.Config{
				Type: generators.TypeInt,
			},
			correct: true,
			version: []int{3, 6},
		},
		{
			name: "int with MaxInt < MinInt",
			config: generators.Config{
				Type:   generators.TypeInt,
				MinInt: 10,
				MaxInt: 4,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "long with missing MinLong and MaxLong",
			config: generators.Config{
				Type: generators.TypeLong,
			},
			correct: true,
			version: []int{3, 6},
		},
		{
			name: "long with MaxLong < MinLong",
			config: generators.Config{
				Type:    generators.TypeLong,
				MinLong: 10,
				MaxLong: 4,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "double with missing MinDouble and MaxDouble",
			config: generators.Config{
				Type: generators.TypeDouble,
			},
			correct: true,
			version: []int{3, 6},
		},
		{
			name: "double with MaxFloat < MinFloat",
			config: generators.Config{
				Type:      generators.TypeDouble,
				MinDouble: 10,
				MaxDouble: 4,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "array with size < 0 ",
			config: generators.Config{
				Type: generators.TypeArray,
				Size: -1,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "array with invalid content",
			config: generators.Config{
				Type: generators.TypeArray,
				Size: 3,
				ArrayContent: &generators.Config{
					Type:      generators.TypeString,
					MinLength: -1,
				},
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "empty fromArray",
			config: generators.Config{
				Type: generators.TypeFromArray,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "fromArray with invalid BSON values",
			config: generators.Config{
				Type: generators.TypeFromArray,
				In: []interface{}{
					bson.M{
						"_id": bson.ObjectId("aaaa"),
					},
				},
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "binary with invalid minLength",
			config: generators.Config{
				Type:      generators.TypeBinary,
				MinLength: -1,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "binary with incorrect MaxLength",
			config: generators.Config{
				Type:      generators.TypeBinary,
				MinLength: 5,
				MaxLength: 2,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "date with incorrect bounds",
			config: generators.Config{
				Type:      generators.TypeDate,
				StartDate: time.Now(),
				EndDate:   time.Unix(10, 10),
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "constant with invalid BSON value",
			config: generators.Config{
				Type: generators.TypeConstant,
				ConstVal: bson.M{
					"_id": bson.ObjectId("aaaa"),
				},
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "autoincrement generator with no type specified",
			config: generators.Config{
				Type:     generators.TypeAutoincrement,
				AutoType: "",
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "ref generator with invalid generator",
			config: generators.Config{
				Type: generators.TypeRef,
				RefContent: &generators.Config{
					Type:      generators.TypeString,
					MinLength: -1,
				},
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "object generator with invalid generator",
			config: generators.Config{
				Type: generators.TypeObject,
				ObjectContent: map[string]generators.Config{
					"key": {
						Type:      generators.TypeString,
						MinLength: -1,
					},
				},
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name:    "missing type",
			config:  generators.Config{},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "null percentage > 100",
			config: generators.Config{
				Type:           generators.TypeString,
				NullPercentage: 120,
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "unknown faker val",
			config: generators.Config{
				Type:   generators.TypeFaker,
				Method: "unknown",
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "type aggregator",
			config: generators.Config{
				Type:           generators.TypeCountAggregator,
				NullPercentage: 10,
			},
			correct: true,
			version: []int{3, 6},
		},
		{
			name: "unknown type",
			config: generators.Config{
				Type: "unknown",
			},
			correct: false,
			version: []int{3, 6},
		},
		{
			name: "decimal with mongodb 3.2",
			config: generators.Config{
				Type: generators.TypeDecimal,
			},
			correct: false,
			version: []int{3, 2},
		},
		{
			name: "decimal with mongodb 3.6",
			config: generators.Config{
				Type: generators.TypeDecimal,
			},
			correct: true,
			version: []int{3, 4},
		},
	}
	// all possible faker methods
	fakerVal := []string{
		generators.MethodCellPhoneNumber,
		generators.MethodCity,
		generators.MethodCityPrefix,
		generators.MethodCitySuffix,
		generators.MethodCompanyBs,
		generators.MethodCompanyCatchPhrase,
		generators.MethodCompanyName,
		generators.MethodCompanySuffix,
		generators.MethodCountry,
		generators.MethodDomainName,
		generators.MethodDomainSuffix,
		generators.MethodDomainWord,
		generators.MethodEmail,
		generators.MethodFirstName,
		generators.MethodFreeEmail,
		generators.MethodJobTitle,
		generators.MethodLastName,
		generators.MethodName,
		generators.MethodNamePrefix,
		generators.MethodNameSuffix,
		generators.MethodPhoneNumber,
		generators.MethodPostCode,
		generators.MethodSafeEmail,
		generators.MethodSecondaryAddress,
		generators.MethodState,
		generators.MethodStateAbbr,
		generators.MethodStreetAddress,
		generators.MethodStreetName,
		generators.MethodStreetSuffix,
		generators.MethodURL,
		generators.MethodUserName,
	}

	for _, f := range fakerVal {
		newGeneratorTests = append(newGeneratorTests, testCase{
			name: fmt.Sprintf(`faker generator with method "%s"`, f),
			config: generators.Config{
				Type:   generators.TypeFaker,
				Method: f,
			},
			correct: true,
			version: []int{3, 6},
		})
	}

	ci := generators.NewCollInfo(100, nil, defaultSeed, map[int][][]byte{}, map[int]byte{})

	for _, tt := range newGeneratorTests {
		t.Run(tt.name, func(t *testing.T) {
			ci.Version = tt.version
			var content = map[string]generators.Config{
				"k": tt.config,
			}
			_, err := ci.NewDocumentGenerator(content)
			if tt.correct && err != nil {
				t.Errorf("expected no error for generator with config %v \nbut got \n%v", tt.config, err)
			}
			if !tt.correct && err == nil {
				t.Errorf("expected no error for generator with config %v", tt.config)
			}
		})
	}
}

func TestNewGeneratorFromMap(t *testing.T) {

	contentList := loadCollConfig(t, "ref.json")

	generatorFromMapTests := []struct {
		name         string
		config       map[string]generators.Config
		correct      bool
		nbGenerators int
	}{
		{
			name: "invalid generator",
			config: map[string]generators.Config{
				"key": {
					Type:      generators.TypeString,
					MinLength: -1,
				},
			},
			correct:      false,
			nbGenerators: 0,
		}, {
			name:         "ref.json[0]",
			config:       contentList[0],
			correct:      true,
			nbGenerators: 14,
		}, {
			name:         "ref.json[1]",
			config:       contentList[1],
			correct:      true,
			nbGenerators: 3,
		},
	}

	ci := generators.NewCollInfo(1, []int{3, 4}, defaultSeed, map[int][][]byte{}, map[int]byte{})

	for _, tt := range generatorFromMapTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ci.NewDocumentGenerator(tt.config)
			if tt.correct {
				if err != nil {
					t.Errorf("expected no error for config %v \n%v", tt.config, err)
				}
			} else if !tt.correct && err == nil {
				t.Errorf("expected an error for config %v but got none", tt.config)
			}
		})
	}
}

func loadCollConfig(t *testing.T, filename string) []map[string]generators.Config {
	bytes, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		t.Error(err)
	}
	var cc []struct {
		Content map[string]generators.Config `json:"content"`
	}
	err = json.Unmarshal(bytes, &cc)
	if err != nil {
		t.Error(err)
	}
	list := make([]map[string]generators.Config, 0, len(cc))
	for _, c := range cc {
		list = append(list, c.Content)
	}
	return list
}

func BenchmarkGeneratorAll(b *testing.B) {

	contentList := loadCollConfig(nil, "ref.json")

	ci := generators.NewCollInfo(1000, []int{3, 2}, defaultSeed, map[int][][]byte{}, map[int]byte{})
	docGenerator, err := ci.NewDocumentGenerator(contentList[0])
	if err != nil {
		b.Fail()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docGenerator.Generate()
	}
}
