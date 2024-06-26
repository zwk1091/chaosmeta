/*
 * Copyright 2022-2023 Chaos Meta Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package experiment

import (
	"chaosmeta-platform/pkg/service/experiment"
)

type CreateExperimentResponse struct {
	UUID string `json:"uuid"`
}

type GetExperimentResponse struct {
	Experiment experiment.ExperimentGet `json:"experiments"`
}

type ExperimentListResponse struct {
	Page        int                        `json:"page"`
	PageSize    int                        `json:"pageSize"`
	Total       int64                      `json:"total"`
	Experiments []experiment.ExperimentGet `json:"experiments"`
}
