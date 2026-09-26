# An administrator is imported by the object id the Pritunl API addresses it
# with.
terraform import pritunl_administrator.ci 6a9b8b734f5041e3afd7a11b

# A username works as well, which is what makes adopting the account the
# instance was seeded with possible without looking its object id up through
# the API first.
terraform import pritunl_administrator.default pritunl

# The password is never read back from the API, so it is not populated by the
# import and the first plan afterwards reports it as a change. That change is a
# write of the configured password, not a reset: the account keeps the password
# it has until one is actually applied.
