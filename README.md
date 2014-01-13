iobify
======

From the raw version and a normalized version of a text, creates a raw version
where sentence and token boundaries are explicitly marked.

Motivation
----------

Many text corpora are distributed in two formats: one "raw" version that
preserves the text as it was found on the Web or in a newspaper archive, and a
"normalized" version that explicitly marks sentence and token boundaries by
putting spaces around all tokens and newlines around all sentences, but also
adds other normalizations concerning e.g. spelling and punctuation. iobify is a
tool for "denormalizing" the text, i.e. keeping the explicit sentence and token
boundaries but throwing away other normalizations. For example:

    raw text:                      Un po' di fresco.
    (erroneously) normalized text: Un pò di fresco .
    iobify output:                 Un po' di fresco .

The purpose is to convert corpora into training data for statistical boundary
detectors such as Elephant [1].

Building
--------

Install Go [2], put the iobify directory into your Go src directory, cd into the
iobify directory and type go install.

Output format
-------------

iobify outputs one character of the raw text a line, with the first
space-separated column containing the character's Unicode code point in
decimal format and the second column containing the character

* S for the first character of the first token of a sentence
* T for the first character of other tokens
* I for other characters that are part of tokens
* O for characters that are not part of tokens (such as whitespace)

References
----------

[1] http://gmb.let.rug.nl/elephant/about.php
[2] http://golang.org/
