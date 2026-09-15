import { Container } from "@maxhub/max-ui";
import '../../styles/Info.css';
import Header from '../../components/Header';
import Footer from '../../components/Footer';
import InfoBlock from '../../components/InfoBlock';

const InfoResident = () => {
    const info = [
        {
            id: 1,
            title: 'Правила пользования микроволновой печью',
            content: 'правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила'
        },
        {
            id: 2,
            title: 'Запрет курения',
            content: 'правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила'
        },
        {
            id: 3,
            title: 'Материальная ответственность',
            content: 'правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила'
        },
        {
            id: 4,
            title: 'Материальная ответственность',
            content: 'правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила'
        },
        {
            id: 5,
            title: 'Материальная ответственность',
            content: 'правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила правила'
        },
    ]
    return (
        <div style={{ minHeight: '100vh' }}>
            <Header />

            <Container>
                <div className='info_list'>
                    {info.map(item => (
                        <InfoBlock key={item.id} content={item.content} title={item.title} />
                    ))}
                </div>
            </Container>
            <Footer />
        </div>
    );
};

export default InfoResident;
