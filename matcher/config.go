/*
Copyright 2016 Palantir Technologies, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package matcher

type namesPathsMatchers struct {
	names Matcher
	paths Matcher
}

type Exclude namesPathsMatchers

func (e *Exclude) Matcher() Matcher {
	return Any(e.names, e.paths)
}

type ExcludeCfg struct {
	Names []string `yaml:"exclude-names" json:"exclude-names"`
	Paths []string `yaml:"exclude-paths" json:"exclude-paths"`
}

func (c *ExcludeCfg) Add(cfg ExcludeCfg) {
	c.Names = append(c.Names, cfg.Names...)
	c.Paths = append(c.Paths, cfg.Paths...)
}

func (c *ExcludeCfg) Exclude() Exclude {
	return Exclude(namesPathsMatchers{
		names: Name(c.Names...),
		paths: Path(c.Paths...),
	})
}

type Include namesPathsMatchers

func (i *Include) Matcher() Matcher {
	return Not(Any(i.names, i.paths))
}

type IncludeCfg struct {
	Names []string `yaml:"include-names" json:"include-names"`
	Paths []string `yaml:"include-paths" json:"include-paths"`
}

func (c *IncludeCfg) Add(cfg IncludeCfg) {
	c.Names = append(c.Names, cfg.Names...)
	c.Paths = append(c.Paths, cfg.Paths...)
}

func (c *IncludeCfg) Include() Include {
	return Include(namesPathsMatchers{
		names: Name(c.Names...),
		paths: Path(c.Paths...),
	})
}
