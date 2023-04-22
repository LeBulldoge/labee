<div align="center">

# labee
Buzz around your files using labels! 🐝

</div>

Labee is a command-line tool that can be used to attach labels to your files.
This allows for grouping related files together, easy access and integration with your favorite CLI tools.

```sh
labee add -l TODO item.txt                  # add a file to the storage and attach the label 'TODO' to it
labee query file i txt                      # find a file using keywords
labee edit label --color "#00FF00" TODO     # modify a label (add a color to it)
labee find --interactive                    # open an interactive view of all files using fzf
labee find -l TODO -n '*.txt' | xargs nvim  # find text files that the label 'TODO' is attached to
                                            # then open the selected files in neovim
```
