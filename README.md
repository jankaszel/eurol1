# eurol1

eurol1 is a tool for post-processing parts of the [Europarl parallel corpus](https://www.statmt.org/europarl/index.html)
and filtering the aligned sentences for their metadata. As of now, the purpose of this tool is to filter out sentences
of a particular original language, since the aligned sentences may have been spoken in another language originally. This
doesn't properly identify the speakers' L1, but comes closer.

## Usage

Say you want to filter all sentences of the parallel Spanish-English corpus that have been originally said in Spanish.
Your directory may look like this, where the former two files contain the aligned sentences in each language and
the `txt` folder contains the corpus source (which includes metadata):

```
europarl-v7.es-en.es
europarl-v7.es-en.en
txt/
    es/
        ...
```

Now, run eurol1:

```bash
$ eurol1 ./europarl-v7.es-en.es ./europarl-v7.es-en.en ./txt/es es-en.filtered.json
```
