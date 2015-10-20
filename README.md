# STEGOsaurus

STEGOsaurus is a dirt simple template evaluation engine.

It stands for Static Template Evaluator in GO.  It processes a directory of templates using the `html/template` package and writes the output out.

Here is what it does:

* From `.` loads an optional context.yml file to be used as template context.
* Walks all files in `./templates` looking for `_foo.tmpl`.  It then parses and uses this as a template named `foo`.  These are "base templates"
* Walks all files in `./templates`.
  * Each file ending in `.tmpl` (but without a leading `_`) is evaluated.  The base templates are available.  If the file begins with a `---` delimited set of lines, those are loaded as a yaml file and merged into the context for this file.
  * All other non-template files are copied verbatim.
* Writes the results into `./output`

## TODO

![I should write some tests](http://i.imgur.com/sVXEH1d.jpg)