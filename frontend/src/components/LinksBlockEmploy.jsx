import { LuUsersRound, LuHouse, LuMessageSquareText } from "react-icons/lu"
import { Button, CellList, CellSimple } from "@maxhub/max-ui";

const LinksBlockEmploy = () => {
    const links = [
        { id: 1, subtitle: 'Беседа этажа', title: '1 Этаж', before: <LuUsersRound /> },
        { id: 2, subtitle: 'Беседа общежития', title: 'Общежитие №1', before: <LuHouse /> },
        { id: 3, subtitle: 'Беседа студ совета', title: 'Студ совет', before: <LuMessageSquareText /> },
    ];
    return (
        <>
            <CellList className="mt-1" filled mode="island">
                {links.map((link) => (
                    <CellSimple
                        key={link.id}
                        title={link.title}
                        subtitle={link.subtitle}
                        before={link.before}
                        showChevron>
                    </CellSimple>
                ))}
            </CellList>

        </>
    );
}

export default LinksBlockEmploy;