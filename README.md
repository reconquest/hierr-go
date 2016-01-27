# Hierarchical errors made right

Hate seeing `error: exit status 128` in the output of programs without actual
explanation what is going wrong?

Or, maybe, you're more advanced programming, and use errors concatenation?

```
can't pull remote 'origin': can't run git fetch 'origin' 'refs/tokens/*:refs/tokens/*': exit status 128
```

Better, but still unreadable.

# hierr

Transform error reports into hierarchy:

```
can't pull remote 'origin'
└─ can't run git fetch 'origin' 'refs/tokens/*:refs/tokens/*'
   └─ exit status 128
```

To use hierarchy error reporting, just convert `fmt.Errorf` calls:

```
return fmt.Errorf("can't pull remote '%s': %s", remote, err)
```

→

```
return hierr.Errorf(err, "can't pull remote '%s'", remote)
```

Docs: https://godoc.org/github.com/seletskiy/hierr
