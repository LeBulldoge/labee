<div align="center">

# labee
Buzz around your files using labels! 🐝

</div>

Introducing `labee`, a file labelling tool written in golang!
With it's simple command-line interface, you can easily attach labels to your files and organize them based on their relationships. This makes it easy for you to access your files and seamlessly integrate them with your favorite CLI tools.

Here are some sample commands that showcase labee's use cases:
```sh
labee add -l TODO item.txt                  # Add the file 'item.txt' to the storage and attach the label 'TODO' to it
labee edit --color '#00FF00' TODO           # Change the label 'TODO' to include the color '#00FF00'
labee find --interactive                    # Open an interactive view of all files inside fzf
labee find -l TODO -n '*.txt' | xargs nvim  # Find all text files with the label 'TODO' attached and open them in neovim
```
