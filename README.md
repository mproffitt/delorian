# `delorian`

Delorian is a TUI for FluxCD at the repo level, designed to interact with the
`flux build` and `flux diff` commands.

> [!Caution]
> At present this is still a work in progress, so some functionality
> may not work, and some bugs and crashes are expected.

To use the UI, enter a gitops repository and run `delorian` - there are
currently no args to the program so don't worry about that (for now).

`delorian` will try and discover details about flux kustomizations, detect
clusters and render manifests.

## Install

```bash
go build . &&
install -m 755 delorian ~/bin/delorian
```

## Usage

Select a flux kustomization in the left menu. Hit `<TAB>` to switch between
the menu and the view area. `;` and `:` switch between tabs

- Kustomize tab shows the rendered Flux kustomization
- Source tab shows the git repository source (if available)
- Flux Build tab runs `flux build` against your current kubernetes context
- Flux Diff runs `flux diff` against your current kubernetes context and
  parses the output.

On the Flux Build pane, you can filter the output using `yq` filters

On the diff pane, you can show / hide parts of the diff by using the
checkboxes at the top.

More documentation to follow
