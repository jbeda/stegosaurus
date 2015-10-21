# STEGOsaurus

STEGOsaurus is a dirt simple template evaluation engine.

It stands for Static Template Evaluator in GO.  It processes a directory of templates using the `html/template` package and writes the output out.

```
$ ./stegosaurus -h
Usage of ./stegosaurus:
  -destination string
      The directory for all output.  It will be created if it doesn't exist. (default "output")
  -source string
      The directory to walk for source templates and files. (default "templates")
```

Here is what it does:

* From `.` loads an optional `context.yml` file to be used as template context.
* Walks all files in `<source>` looking for `_foo.tmpl`.  It then parses and uses this as a template named `foo`.  These are "base templates"
* Walks all files in `<source>`.
  * For each file ending in `.tmpl` and without a leading `_`:
    * Looks for a set of lines that start and end with `---`. If found, those lines are parsed as a YAML file and merged into a copy of the context.
    * Evaluates the template. All of the base templates are available to be called. The context object `.` can also be used.
  * All other non-template files are copied verbatim.
* Writes the results into `<destination>`

For details on using the golang templating system, see [html/template](https://golang.org/pkg/html/template/) and [text/template](https://golang.org/pkg/text/template/).

## Example

This example is in the `example` directory.

```
.
├── context.yml
├── templates
│   ├── _base.tmpl
│   ├── css
│   │   └── site.css
│   └── index.html.tmpl
└── watch.sh
```

### Running

```console
$ cd example
$ go run ../stegosaurus.go
Loading template file: templates/_base.tmpl
Copying templates/css/site.css to output/css/site.css
Rendering templates/index.html.tmpl to output/index.html
```

You can also set up a watch with fswatch to re-run the template system on every save.  There is a script called `watch.sh` to do that:

```bash
fswatch -o context.yml templates | xargs -tn 1 -I {} go run ../stegosaurus.go
```

### `context.yml`
This sets data that will be available to every template.

```yaml
title: "Default title"
year: 2015
author: "Joe Beda"
navCurrent: ""
navbar:
  - text: home
    url: "/"
  - text: about
    url: about.html
  - text: "@jbeda"
    url: "https://twitter.com/jbeda"
```

### `_base.tmpl`
This is a simple HTML5 template page.  Data like the title and the copyright info are taken from the context.  It defines a "subtemplate" called `content` that is empty.  Because it is empty, users can override/redefine it in other templates.

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{.title}}</title>
    <link rel="stylesheet" href="css/site.css">
  </head>
  <body>
    <ul class=nav>
    {{range .navbar}}
      <li {{if eq $.navCurrent .text}}class="active"{{end}}><a href="{{.url}}">{{.text}}</a></li>
    {{end}}
    </ul>
    <h1>{{.title}}</h1>
    {{template "content" . }}
    <div class="copyright">&copy; {{.year}} {{.author}}</div>
  </body>
</html>

{{define "content"}}{{end}}
```

### `css/site.css`

This is a static file that is copied verbatim into the output.  The directory will be preserved/created in the output structure.  Here is a lame example:

```css
.copyright {
  font-size: 0.5rem;
}

ul.nav {
  padding: 0;
}

ul.nav li {
  font-family: sans-serif;
  display:inline-block;
  padding: 10px;
  background: #bbb;
  border: solid 1px black;
  border-radius: 0.5rem;
}

ul.nav li.active {
  background: #fff;
}
```

### `index.html.tmpl`

This file will be evaulated and writing to `<output>/index.html`.

```html
---
title: My cool site
year: 2020
navCurrent: "home"
---

{{define "content"}}
This is my cool content.
{{end}}

{{template "base" .}}
```

## TODO

![I should write some tests](http://i.imgur.com/sVXEH1d.jpg)