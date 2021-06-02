# Project Page

This is the source of our project page https://nokia.github.io/ntt

Content is written in Markdown and the HTML pages are generated using Jekyll.

When you add content please preview you changes locally. You can install all required tools by running:
	make install-tools

You start the local web-server by running:
	make


## Screen casts - one picture says more than a thousand words

Screen casts are a great way to show users how something works.  Good screen
casts have a consistent look and feel. They should not be too fast, so the user
has time to orient herself. Neither should screen cast too short or too long;
5 to 10 seconds as rule of thumb.


**Console**

The tool `screencast` allows you to create fancy terminal screen casts in SVG
format.  
The svg tool however has some issues with terminals and vim sessions. Help with
with issue is very much appreciated.


**VSCode**

Here are some tipps when recording VS Code sessions:
* _consistent colors_: `Ctrl`,`K` and `Ctrl`, `T`, followed by `dark visual`.
* _consistent size_: `Ctrl`,`Shift`,`P` and `developer tools`. In Javascript
  `console` tab enter: `window.resizeTo(1920, 1080)`
* _screencast mode_: `Ctrl`,`Shift`,`P` and `screencast`.
* _increase font size_: `Ctrl` and `+`

For recording you can use, for example Peek.
