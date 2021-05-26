all:
	bundle exec jekyll serve --livereload

install-tools:
	pip3 install asciinema
	npm install -g svg-term-cli
