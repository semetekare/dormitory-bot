import { Button, CellList, CellSimple } from "@maxhub/max-ui";
import Header from "../../components/Header";
import { LuUsersRound, LuHouse, LuMessageSquareText } from "react-icons/lu"
import Footer from "../../components/Footer";

const ChatLinksResident = () => {
    const links = [
        { id: 1, subtitle: 'Беседа этажа', title: '1 Этаж', before: <LuUsersRound /> },
        { id: 2, subtitle: 'Беседа общежития', title: 'Общежитие №1', before: <LuHouse /> },
        { id: 3, subtitle: 'Беседа студ совета', title: 'Студ совет', before: <LuMessageSquareText /> },
    ];
    return (
        <div style={{ minHeight: '100vh' }}>
            <Header />
            <CellList className="mt-1" filled mode="island">
                {links.map((link) => (
                    <CellSimple
                        key={link.id}
                        title={link.title}
                        subtitle={link.subtitle}
                        before={link.before}
                        showChevron
                        asChild
                    >
                        <a
                            href="https://dev.max.ru"
                            rel="noreferrer"
                            target="_blank"
                        />
                    </CellSimple>
                ))}
            </CellList>
            <Footer />
        </div>
    )
};

export default ChatLinksResident;