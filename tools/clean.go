package tools

import "github.com/galayx-future/billcompare/internal/types"

func CleanUp(diff []types.Diff) []types.Diff {
	result := make([]types.Diff, 0, len(diff))
	for i := range diff {
		if diff[i].SchedulxArrears.IsZero() && diff[i].AlibabaCloudArrears.IsZero() && diff[i].SchedulxCost.IsZero() && diff[i].AlibabaCloudCost.IsZero() {
			continue
		}
		result = append(result, diff[i])
	}
	return result
}
