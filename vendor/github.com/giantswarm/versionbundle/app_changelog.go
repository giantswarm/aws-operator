package versionbundle

type AppsChangelogs map[string][]Changelog

type ReleasesChangelogs map[string]AppsChangelogs
