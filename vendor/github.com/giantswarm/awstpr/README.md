[![CircleCI](https://circleci.com/gh/giantswarm/awstpr.svg?&style=shield&circle-token=752d6ec42799fb9fd56dcf13c142d2f675b0b91f)](https://circleci.com/gh/giantswarm/awstpr)

# awstpr

Specification of the third party objects used to deploy Kubernetes on top of AWS by the "undercloud" Kubernetes running
[aws-operator](https://github.com/giantswarm/aws-operator).

## Getting Project

Clone the git repository: https://github.com/giantswarm/awstpr.git

Check out the latest tag: https://github.com/giantswarm/awstpr/tags

### How to build

This project provides a Makefile, so you can build it by typing:

```
make
```

If you prefer, you may also build it using the standard `go build` command, like:

```
go build github.com/giantswarm/awstpr
```

However, since this project is just a specification used by the other projects, there only goal of building is to check
whether the build is successful. This is just a library which needs to be vendored by the project aiming to use it.

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/awstpr/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the contribution workflow as well as reporting bugs.

## License

awstpr is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
