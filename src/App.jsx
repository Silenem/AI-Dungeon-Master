import { Routes, Route } from 'react-router-dom'
import MyCharacters from './MyCharacters'
import CreateCharacter from './CreateCharacter'
import StorySelect from './StorySelect'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<MyCharacters />} />
      <Route path="/my_characters" element={<MyCharacters />} />
      <Route path="/create_character" element={<CreateCharacter />} />
      <Route path="/stories" element={<StorySelect />} />
    </Routes>
  )
}