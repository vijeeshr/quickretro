import { CategoryDefinition } from '../models/CategoryDefinition'

export const defaultCategories: CategoryDefinition[] = [
  { id: 'col01', text: '', color: 'green', colorClass: 'text-green-500', enabled: true, pos: 1 },
  { id: 'col02', text: '', color: 'red', colorClass: 'text-red-500', enabled: true, pos: 2 },
  { id: 'col03', text: '', color: 'yellow', colorClass: 'text-yellow-500', enabled: true, pos: 3 },
  {
    id: 'col04',
    text: '',
    color: 'fuchsia',
    colorClass: 'text-fuchsia-500',
    enabled: false,
    pos: 4,
  },
  { id: 'col05', text: '', color: 'orange', colorClass: 'text-orange-500', enabled: false, pos: 5 },
]
