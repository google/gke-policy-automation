# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

package gke.rule.cluster.location_test

import future.keywords.if
import data.gke.rule.cluster.location

test_regional if { 
    loc := "europe-central2"
    location.regional(loc)
    not location.zonal(loc)
}

test_zonal if {
    loc := "europe-central2-a"
    location.zonal(loc)
    not location.regional(loc)
}

test_not_regional_nor_zonal if {
    loc := "test"
    not location.regional(loc)
    not location.zonal(loc)
}
