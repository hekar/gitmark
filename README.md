# gitmark

[Goji](https://github.com/zenazn/goji) based server that writes bookmark entries to a git repository
such as [hekar/bookmarks](https://github.com/hekar/bookmarks).

Bookmarks are written in Markdown

## Usage

### Configuration

Create yaml file `~/.gitmarkrc.yaml` (or anything supported by [viper](https://github.com/spf13/viper))

Example (fill in the variables):
```
RepoUrl: "https://<username>:<password>@github.com/<username>/bookmarks.git"
Remote: "origin"
Branch: "master"
MessagePrefix: "Added Bookmark "
UserName: "<your_name>"
UserEmail: "<your_email>"
```

### Add a bookmark

```
curl -X POST --data "title=Google&url=http://google.ca" "http://localhost:8000/bookmark/<username>%252Fbookmarks"
```

## TODO
* [ ] Create a CLI
* [ ] Support more than Markdown (JSON, HTML, etc.)

## License

```
Copyright 2016 Hekar Khani

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```
