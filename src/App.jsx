import { Routes, Route } from 'react-router-dom'
import MyCharacters from './MyCharacters'
import CreateCharacter from './CreateCharacter'
import StorySelect from './StorySelect'
import Register from './auth'
import Profile from './Profile'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<MyCharacters />} />
      <Route path="/my_characters" element={<MyCharacters />} />
      <Route path="/create_character" element={<CreateCharacter />} />
      <Route path="/stories" element={<StorySelect />} />
      <Route path="/auth" element={<Register />} />
      <Route path="/profile" element={<Profile />} />
    </Routes>
  )
}
