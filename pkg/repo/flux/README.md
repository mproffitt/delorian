# Flux repo traversing

The traverse functionality here is quite complicated as it relies on a number of
boundaries

1. Only Flux kustomizations included from deployed paths should be considered.
   Here, bases and overlays are not considered to be deployed paths but included
   paths

2. Included paths may contain flux kustomizations that are patched from deployed
   paths. These should be filtered out or represented in a "templates" or "bases"
   sub-structure

3. Flux kustomizations may be accompanied by kustomize kustomizations. Where a
   kustomize kustomization is found, it should be applied to the flux kustomization
   to apply any patches, etc. However this leads to "flux build" taking place

4. A Kustomize kustomization may point to another kustomize kustomization, creating
   many layers before the final flux kustomization is built

5. Fragments of flux kustomizations may be discovered in patch files. Where this
   is so, these need to be treated differently.

This is kind of a big package and technically needs splitting into sub-modules

- Sidebar Models
  - Kustomization list
  - Git Source list (this doesn't yet exist)
  - Cluster tree

This should leave only repo traversal in this package, which would make it more
manageable
