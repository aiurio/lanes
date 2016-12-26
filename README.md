# IntelliJ Setup Notes

After pulling a dep, you'l have to recreate a symlink `ln -s ./ vendor/src` 
and add the vendor folder to the `Project Libraries` section in `Libraries & Frameworks` > `Go` > `Go Libraries`.

A useful alias can cover both (thrown into my `~/.profile`): 
```
glider () {
        glide get "${@:1}" && ln -s ./ vendor/src
}
```

which then just replaces `glide get xyz` with `glider get xyz`
