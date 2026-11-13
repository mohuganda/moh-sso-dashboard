package geography

import "testing"

func TestAdoptionRejectsAmbiguousLocations(t *testing.T) {
	for _, s := range []snapshot{
		{"regions": {{"id": "one", "name": "North"}, {"id": "two", "name": " north "}}},
		{"facilities": {{"id": "one", "name": "One", "external_id": "same"}, {"id": "two", "name": "Two", "external_id": "same"}}},
	} {
		if _, err := adopt(snapshot{}, s); err == nil {
			t.Fatal("ambiguous reference data was silently merged")
		}
	}
}

func TestMatchingKeepsSameNamedFacilitiesInDifferentDistrictsSeparate(t *testing.T) {
	first := record{"id": "one", "name": "Clinic", "district_id": "north"}
	second := record{"id": "two", "name": "Clinic", "district_id": "south"}
	lookup := index("facilities", []record{first, second})
	found, err := match("facilities", record{"name": "clinic", "district_id": "south"}, lookup, "")
	if err != nil || value(found, "id") != "two" {
		t.Fatalf("wrong district match: %v %v", found, err)
	}
	if _, err := match("facilities", second, lookup, "one"); err == nil {
		t.Fatal("conflicting UUID and natural identity were accepted")
	}
}
