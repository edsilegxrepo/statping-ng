package groups

import (
	"sort"
	"testing"

	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var example = &Group{
	Id:     1,
	Name:   "Example Group",
	Public: null.NewNullBool(true),
	Order:  1,
}

var s1 = &services.Service{
	Name:    "Example Service",
	Public:  null.NewNullBool(true),
	Order:   1,
	GroupId: 1,
}

var s2 = &services.Service{
	Name:    "Example Service 2",
	Public:  null.NewNullBool(true),
	Order:   2,
	GroupId: 1,
}

func TestInit(t *testing.T) {
	// DB setup moved to TestMain in main_test.go
	// This test now just creates the example data
	testDb.Create(&example)
	testDb.Create(&s1)
	testDb.Create(&s2)
}

func TestFind(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	assert.Equal(t, "Example Group", item.Name)
}

func TestFindNonExistent(t *testing.T) {
	item, err := Find(99999)
	assert.Nil(t, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAll(t *testing.T) {
	items := All()
	assert.Len(t, items, 1)
}

func TestValidate(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		g := &Group{Name: ""}
		err := g.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group name is empty")
	})

	t.Run("whitespace only name", func(t *testing.T) {
		g := &Group{Name: "   "}
		err := g.Validate()
		// Current implementation does not trim whitespace, so this passes
		// This documents the current behavior
		assert.Nil(t, err)
	})

	t.Run("valid name", func(t *testing.T) {
		g := &Group{Name: "Valid Group"}
		err := g.Validate()
		assert.Nil(t, err)
	})
}

func TestCreateValidationError(t *testing.T) {
	g := &Group{Name: ""}
	err := g.Create()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group name is empty")
}

func TestUpdateValidationError(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	originalName := item.Name
	item.Name = ""
	err = item.Update()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group name is empty")
	// Restore for subsequent tests
	item.Name = originalName
}

func TestCreate(t *testing.T) {
	example := &Group{
		Name:   "Example 2",
		Public: null.NewNullBool(false),
		Order:  3,
	}
	err := example.Create()
	require.Nil(t, err)
	assert.NotZero(t, example.Id)
	assert.Equal(t, "Example 2", example.Name)
	assert.NotZero(t, example.CreatedAt)
}

func TestUpdate(t *testing.T) {
	item, err := Find(1)
	require.Nil(t, err)
	item.Name = "Updated"
	item.Order = 1
	err = item.Update()
	require.Nil(t, err)
	assert.Equal(t, "Updated", item.Name)
}

func TestSelectGroups(t *testing.T) {
	groups := SelectGroups(true, true)
	assert.Len(t, groups, 2)

	groups = SelectGroups(false, false)
	assert.Len(t, groups, 1)

	groups = SelectGroups(false, true)
	assert.Len(t, groups, 2)

	assert.Equal(t, "Updated", groups[0].Name)
	assert.Equal(t, "Example 2", groups[1].Name)
}

func TestGroupOrder(t *testing.T) {
	t.Run("Len", func(t *testing.T) {
		groups := GroupOrder{
			&Group{Id: 1, Name: "A", Order: 3},
			&Group{Id: 2, Name: "B", Order: 1},
			&Group{Id: 3, Name: "C", Order: 2},
		}
		assert.Equal(t, 3, groups.Len())
	})

	t.Run("Swap", func(t *testing.T) {
		groups := GroupOrder{
			&Group{Id: 1, Name: "A", Order: 3},
			&Group{Id: 2, Name: "B", Order: 1},
		}
		groups.Swap(0, 1)
		assert.Equal(t, "B", groups[0].Name)
		assert.Equal(t, "A", groups[1].Name)
	})

	t.Run("Less", func(t *testing.T) {
		groups := GroupOrder{
			&Group{Id: 1, Name: "A", Order: 3},
			&Group{Id: 2, Name: "B", Order: 1},
		}
		assert.False(t, groups.Less(0, 1)) // 3 < 1 is false
		assert.True(t, groups.Less(1, 0))  // 1 < 3 is true
	})

	t.Run("sort integration", func(t *testing.T) {
		groups := GroupOrder{
			&Group{Id: 1, Name: "Third", Order: 3},
			&Group{Id: 2, Name: "First", Order: 1},
			&Group{Id: 3, Name: "Second", Order: 2},
		}
		sort.Sort(groups)
		assert.Equal(t, "First", groups[0].Name)
		assert.Equal(t, "Second", groups[1].Name)
		assert.Equal(t, "Third", groups[2].Name)
	})

	t.Run("empty slice", func(t *testing.T) {
		groups := GroupOrder{}
		assert.Equal(t, 0, groups.Len())
		sort.Sort(groups) // Should not panic
	})

	t.Run("single element", func(t *testing.T) {
		groups := GroupOrder{&Group{Id: 1, Name: "Only", Order: 1}}
		assert.Equal(t, 1, groups.Len())
		sort.Sort(groups)
		assert.Equal(t, "Only", groups[0].Name)
	})

	t.Run("same order values", func(t *testing.T) {
		groups := GroupOrder{
			&Group{Id: 1, Name: "A", Order: 1},
			&Group{Id: 2, Name: "B", Order: 1},
		}
		assert.False(t, groups.Less(0, 1)) // 1 < 1 is false
		assert.False(t, groups.Less(1, 0)) // 1 < 1 is false
	})
}

func TestDelete(t *testing.T) {
	all := All()
	assert.Len(t, all, 2)

	item, err := Find(1)
	require.Nil(t, err)

	err = item.Delete()
	require.Nil(t, err)

	all = All()
	assert.Len(t, all, 1)
}

func TestSamples(t *testing.T) {
	require.Nil(t, Samples())
	assert.Len(t, All(), 4)
}

func TestSelectGroupsPublicFiltering(t *testing.T) {
	// At this point we have 4 groups from Samples:
	// - Main Services (public, order 2)
	// - Linked Services (public, order 1)
	// - Private Services (private, order 3)
	// - Example 2 (private, order 3) from TestCreate

	t.Run("includeAll returns all groups", func(t *testing.T) {
		groups := SelectGroups(true, false)
		assert.Len(t, groups, 4)
	})

	t.Run("unauthenticated sees only public groups", func(t *testing.T) {
		groups := SelectGroups(false, false)
		// Should only include public groups
		for _, g := range groups {
			assert.True(t, g.Public.Bool, "unauthenticated user should only see public groups")
		}
	})

	t.Run("authenticated sees all groups", func(t *testing.T) {
		groups := SelectGroups(false, true)
		assert.Len(t, groups, 4)
	})

	t.Run("groups are sorted by order", func(t *testing.T) {
		groups := SelectGroups(true, true)
		for i := 0; i < len(groups)-1; i++ {
			assert.LessOrEqual(t, groups[i].Order, groups[i+1].Order,
				"groups should be sorted by Order field")
		}
	})
}

func TestClose(t *testing.T) {
	assert.Nil(t, db.Close())
}
