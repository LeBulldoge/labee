# Labee
Buzz around your files using labels!

Labee is a command-line tool that can be used to attach labels to your files.
Query the storage for the files with label names for fast access and easy integration with other tools via pipelines.

```sh
labee add -l TODO item.txt                  # add a file to the storage and attach the label 'TODO' to it
labee query file i txt                      # find a file using keywords
labee edit label --color "#00FF00" TODO     # modify a label (add a color to it)
labee query label --all --interactive       # open an interactive view of all labels using fzf
labee query label -i TODO | nvim            # find files that the label is attached to in interactive mode
                                            # then open the selected files in neovim
```
