### A GO autogenerated files cleaner

just install with go install and use it like gencode-remover ./...

also you can use it like

gencode-remover --output ./... | xargs git rm -f
