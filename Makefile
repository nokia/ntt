all:
	bundle exec jekyll serve

install-tools:
	pip3 install asciinema
	npm install -g svg-term-cli
