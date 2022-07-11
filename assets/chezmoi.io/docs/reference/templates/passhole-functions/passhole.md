# passhole(ph) *command*
`ph` returns structured data from a [Keepass](https://keepass.info/) database. Users can add, modify, delete, search entries in the database by using ph's commands: "list,ls,add,remove,rm,move,mv,show,edit,type,init,grep,dump,info,kill"

In passhole function, we are using show commands twice to retrieve username and password:
`ph show --field username *entry name*`
`ph show --field password *entry name*`
