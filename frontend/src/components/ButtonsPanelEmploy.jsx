import { Flex, ToolButton } from "@maxhub/max-ui";
import { LuLink, LuBadgeInfo, LuDoorOpen, LuWashingMachine, LuCalendarCheck2 } from 'react-icons/lu'


const ButtonsPanelEmploy = () => (
    <>
        <Flex gapX={12} gapY={12} align='center' justify='center'>
            <ToolButton asChild icon={<LuWashingMachine style={{ fontSize: '2em' }} />}>
                <a href='/laundry_employ'>Стирка</a>
            </ToolButton>
            <ToolButton asChild icon={<LuDoorOpen style={{ fontSize: '2em' }} />}>
                <a href='/room_employ'>Комнаты</a>
            </ToolButton>
            <ToolButton asChild icon={<LuCalendarCheck2 style={{ fontSize: '2em' }} />}>
                <a href='/cleaning_employ'>Дежурства</a>
            </ToolButton>
        </Flex>
        <Flex gapX={12} gapY={12} align='center' justify='center'>
            <ToolButton asChild icon={<LuBadgeInfo style={{ fontSize: '2em' }} />}>
                <a href='/info_employ'>Информация</a>
            </ToolButton>
            <ToolButton asChild icon={<LuLink style={{ fontSize: '2em' }} />}>
                <a href='/chat-links_employ'>Ссылки на чаты</a>
            </ToolButton>
        </Flex>
    </>
);
export default ButtonsPanelEmploy;
