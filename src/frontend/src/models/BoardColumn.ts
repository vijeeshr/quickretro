export interface BoardColumn {
  id: string
  text: string
  isDefault: boolean // Used to identify if Board creator entered custom value for "text". Useful during multi-lang translation.
  color: string
  pos: number
}
