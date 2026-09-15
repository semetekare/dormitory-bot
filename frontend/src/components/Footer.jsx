import '../styles/Footer.css';

const Footer = () => {
    return (
        <footer className='footer'>
            <div className='footer_block'>
                <div className='footer_content_minimal'>
                    <span>БОТ ОБЩЕЖИТИЯ {new Date().getFullYear()}</span>
                    <span>Поддержка: 0202@irgups.ru</span>
                </div>
            </div>
        </footer>
    );
};

export default Footer;