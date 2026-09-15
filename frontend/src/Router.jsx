import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import RoomInfo from './pages/Employ/RoomInfo';

// ── Static imports — Vite cannot code-split template-string dynamic imports ──
// Resident pages
import HomeResident from './pages/Resident/HomeResident';
import ChatLinksResident from './pages/Resident/ChatLinksResident';
import InfoResident from './pages/Resident/InfoResident';
import RoomResident from './pages/Resident/RoomResident';
import LaundryResident from './pages/Resident/LaundryResident';
import CleaningResident from './pages/Resident/CleaningResident';

// Employ pages
import HomeEmploy from './pages/Employ/HomeEmploy';
import ChatLinksEmploy from './pages/Employ/ChatLinksEmploy';
import InfoEmploy from './pages/Employ/InfoEmploy';
import RoomEmploy from './pages/Employ/RoomEmploy';
import LaundryEmploy from './pages/Employ/LaundryEmploy';
import CleaningEmploy from './pages/Employ/CleaningEmploy';

// StudCouncil pages
import HomeStudentCouncil from './pages/StudCouncil/HomeStudentCouncil';
import ChatLinksStudentCouncil from './pages/StudCouncil/ChatLinksStudentCouncil';
import InfoStudentCouncil from './pages/StudCouncil/InfoStudentCouncil';
import RoomStudentCouncil from './pages/StudCouncil/RoomStudentCouncil';
import LaundryStudentCouncil from './pages/StudCouncil/LaundryStudentCouncil';
import CleaningStudentCouncil from './pages/StudCouncil/CleaningStudentCouncil';

const pagesByRole = {
  resident: {
    Home: HomeResident,
    ChatLinks: ChatLinksResident,
    Info: InfoResident,
    Room: RoomResident,
    Laundry: LaundryResident,
    Cleaning: CleaningResident,
  },
  employ: {
    Home: HomeEmploy,
    ChatLinks: ChatLinksEmploy,
    Info: InfoEmploy,
    Room: RoomEmploy,
    Laundry: LaundryEmploy,
    Cleaning: CleaningEmploy,
  },
  stud_council: {
    Home: HomeStudentCouncil,
    ChatLinks: ChatLinksStudentCouncil,
    Info: InfoStudentCouncil,
    Room: RoomStudentCouncil,
    Laundry: LaundryStudentCouncil,
    Cleaning: CleaningStudentCouncil,
  },
};

function Router({ role }) {
  const pages = pagesByRole[role] || pagesByRole.resident;

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to={`/home_${role}`} replace />} />
        <Route path={`/home_${role}`} element={<pages.Home />} />
        <Route path={`/chat-links_${role}`} element={<pages.ChatLinks />} />
        <Route path={`/info_${role}`} element={<pages.Info />} />
        <Route path={`/room_${role}`} element={<pages.Room />} />
        <Route path={`/laundry_${role}`} element={<pages.Laundry />} />
        <Route path={`/cleaning_${role}`} element={<pages.Cleaning />} />
        <Route path="/cleaning_employ/room_info" element={<RoomInfo />} />
      </Routes>
    </BrowserRouter>
  );
}

export default Router;
