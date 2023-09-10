# Application order

chezmoi is deterministic in its order of application. The order is:

1. Read the source state.
2. Read the destination state.
3. Compute the target state.
4. Run `run_before_` scripts in alphabetical order.
5. Update entries in the target state (files, directories, externals, scripts,
   symlinks, etc.) in alphabetical order of their target name. Directories
   (including those created by externals) are updated before the files they
   contain.
6. Run `run_after_` scripts in alphabetical order.

Target names are considered after all attributes are stripped.

!!! example

    Given `create_alpha` and `modify_dot_beta` in the source state, `.beta`
    will be updated before `alpha` because `.beta` sorts before `alpha`.

chezmoi assumes that the source or destination states are not modified while
chezmoi is being executed. This assumption permits significant performance
improvements, including allowing chezmoi to only read files from the source and
destination states if they are needed to compute the target state.

chezmoi's behavior when the above assumptions are violated is undefined. For
example, using a `run_before_` script to update files in the source or
destination states violates the assumption that the source and destination
states do not change while chezmoi is running.

!!! note

    External sources are updated during the update phase; it is inadvisable for
    a `run_before_` script to depend on an external applied *during* the update
    phase. `run_after_` scripts may freely depend on externals.
