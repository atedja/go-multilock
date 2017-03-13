# Multilock

Multilock allows you to obtain multiple locks without deadlock. It also uses
strings as locks, which allows multiple goroutines to synchronize independently
without having to share common mutex objects.

One common application is to use an external id (e.g. IDs from a database)
as the lock, and thereby preventing multiple goroutines from potentially
reading/writing to the same row in the database.
