import { useLocation } from 'react-router';
import IrgupsLine from '../assets/svg/irgups-line-green.svg';
import '../styles/Header.css';
import BackButton from './BackButton';

const Header = () => {
  const location = useLocation();
  const homePaths = ['/home_resident', '/home_employ', '/home_stud_council'];
  const isHome = !homePaths.includes(location.pathname);

  return (
    <header className='header'>
      <h1 className="header-title">БОТ ОБЩЕЖИТИЯ</h1>
      <img className="irgups_line" src={IrgupsLine} alt="" />
      <img className="irgups_line_2" src={IrgupsLine} alt="" />
      <img className="irgups_line_3" src={IrgupsLine} alt="" />
      {isHome && (
        <div className='header-button'>
          <BackButton />
        </div>
      )}
    </header>
  );
};

export default Header;